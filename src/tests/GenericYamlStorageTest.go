package tests

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"reflect"
	"rol/domain"
	"rol/infrastructure"
	"strings"
)

//GenericYamlStorageTest is a generic structure for testing yaml storage
type GenericYamlStorageTest[TemplateType domain.DeviceTemplate] struct {
	Storage        *infrastructure.YamlGenericTemplateStorage[TemplateType]
	Context        context.Context
	DirName        string
	TemplatesCount int
}

//NewGenericYamlStorageTest constructor for GenericYamlStorageTest
func NewGenericYamlStorageTest[TemplateType domain.DeviceTemplate](storage *infrastructure.YamlGenericTemplateStorage[TemplateType], dirName string, templatesCount int) *GenericYamlStorageTest[TemplateType] {
	if templatesCount < 2 {
		templatesCount = 2
	}
	return &GenericYamlStorageTest[TemplateType]{
		Storage:        storage,
		Context:        context.TODO(),
		DirName:        dirName,
		TemplatesCount: templatesCount,
	}
}

func createXTemplatesForTest(x int) error {
	err := os.Mkdir("../templates/testTemplates", 0777)
	if err != nil {
		return fmt.Errorf("creating dir failed: %s", err)
	}
	for i := 1; i <= x; i++ {
		template := domain.DeviceTemplate{
			Name:         fmt.Sprintf("AutoTesting_%d", i),
			Model:        fmt.Sprintf("AutoTesting_%d", i),
			Manufacturer: "Manufacturer",
			Description:  "Description",
			CPUCount:     i,
			CPUModel:     "CPUModel",
			RAM:          i,
			NetworkInterfaces: []domain.DeviceTemplateNetworkInterface{{
				Name:       "Name",
				NetBoot:    false,
				POEIn:      false,
				Management: false,
			}},
			Control: domain.DeviceTemplateControlDesc{
				Emergency: "Emergency",
				Power:     "Power",
				NextBoot:  "NextBoot",
			},
			DiscBootStages: []domain.BootStageTemplate{{
				Name:        "Name",
				Description: "Description",
				Action:      "Action",
				Files: []domain.BootStageTemplateFile{{
					ExistingFileName: "ExistingFileName",
					VirtualFileName:  "VirtualFileName",
				}},
			}},
			NetBootStages: []domain.BootStageTemplate{{
				Name:        "Name",
				Description: "Description",
				Action:      "Action",
				Files: []domain.BootStageTemplateFile{{
					ExistingFileName: "ExistingFileName",
					VirtualFileName:  "VirtualFileName",
				}},
			}},
			USBBootStages: []domain.BootStageTemplate{{
				Name:        "Name",
				Description: "Description",
				Action:      "Action",
				Files: []domain.BootStageTemplateFile{{
					ExistingFileName: "ExistingFileName",
					VirtualFileName:  "VirtualFileName",
				}},
			}},
		}
		if i == 2 {
			template.Description = "ValueForSearch"
		}
		yamlData, err := yaml.Marshal(&template)
		if err != nil {
			return fmt.Errorf("yaml marshal failed: %s", err)
		}
		fileName := fmt.Sprintf("../templates/testTemplates/AutoTesting_%d.yml", i)
		err = ioutil.WriteFile(fileName, yamlData, 0777)
		if err != nil {
			return fmt.Errorf("create yaml file failed: %s", err)
		}
	}
	return nil
}

func (g *GenericYamlStorageTest[TemplateType]) GetByNameTest(fileName string) error {
	nameSlice := strings.Split(fileName, ".")
	name := nameSlice[0]
	template, err := g.Storage.GetByName(g.Context, fileName)
	if err != nil {
		return fmt.Errorf("get by name failed: %s", err)
	}

	obtainedName := reflect.ValueOf(*template).FieldByName("Name").String()

	if obtainedName != name {
		return fmt.Errorf("unexpected name %s, expect %s", obtainedName, name)
	}
	return nil
}

func (g *GenericYamlStorageTest[TemplateType]) GetListTest() error {
	templatesArr, err := g.Storage.GetList(g.Context, "", "", 1, g.TemplatesCount, nil)
	if err != nil {
		return fmt.Errorf("get list failed:  %s", err)
	}
	if len(*templatesArr) != g.TemplatesCount {
		return fmt.Errorf("array length %d, expect %d", len(*templatesArr), g.TemplatesCount)
	}
	return nil
}

func (g *GenericYamlStorageTest[TemplateType]) PaginationTest(page, pageSize int) error {
	templatesArrFirstPage, err := g.Storage.GetList(g.Context, "CPUCount", "asc", page, pageSize, nil)
	if err != nil {
		return fmt.Errorf("get list failed: %s", err)
	}
	if len(*templatesArrFirstPage) != pageSize {
		return fmt.Errorf("array length on %d page %d, expect %d", page, len(*templatesArrFirstPage), pageSize)
	}
	templatesArrSecondPage, err := g.Storage.GetList(g.Context, "CPUCount", "asc", page+1, pageSize, nil)
	if err != nil {
		return fmt.Errorf("get list failed: %s", err)
	}
	if len(*templatesArrSecondPage) != pageSize {
		return fmt.Errorf("array length on next page %d, expect %d", len(*templatesArrSecondPage), pageSize)
	}

	firstPageValue := reflect.ValueOf(*templatesArrFirstPage).Index(0)
	secondPageValue := reflect.ValueOf(*templatesArrSecondPage).Index(0)

	firstPageValueName := firstPageValue.FieldByName("Name").String()
	secondPageValueName := secondPageValue.FieldByName("Name").String()
	if firstPageValueName == secondPageValueName {
		return fmt.Errorf("pagination failed: got same element on second page with Name: %s", firstPageValueName)
	}
	firstPageValueCPU := firstPageValue.FieldByName("CPUCount").Int()
	secondPageValueCPU := secondPageValue.FieldByName("CPUCount").Int()

	if secondPageValueCPU-int64(pageSize) != firstPageValueCPU {
		return fmt.Errorf("pagination failed: unexpected element on second page")
	}

	return nil
}

func (g *GenericYamlStorageTest[TemplateType]) SortTest() error {
	templatesArr, err := g.Storage.GetList(g.Context, "CPUCount", "asc", 1, g.TemplatesCount, nil)
	if err != nil {
		return fmt.Errorf("get list failed: %s", err)
	}
	if len(*templatesArr) != g.TemplatesCount {
		return fmt.Errorf("array length %d, expect %d", len(*templatesArr), g.TemplatesCount)
	}
	index := g.TemplatesCount / 2
	name := reflect.ValueOf(*templatesArr).Index(index - 1).FieldByName("Name").String()

	if name != fmt.Sprintf("AutoTesting_%d", index) {
		return fmt.Errorf("sort failed: got %s name, expect AutoTesting_%d", name, index)
	}
	return nil
}

func (g *GenericYamlStorageTest[TemplateType]) FilterTest() error {
	queryBuilder := g.Storage.NewQueryBuilder(g.Context)
	queryBuilder.
		Where("CPUCount", ">", g.TemplatesCount/2).
		Where("CPUCount", "<", g.TemplatesCount)
	templatesArr, err := g.Storage.GetList(g.Context, "", "", 1, g.TemplatesCount, queryBuilder)
	if err != nil {
		return fmt.Errorf("get list failed: %s", err)
	}
	var expectedCount int
	if g.TemplatesCount%2 == 0 {
		expectedCount = g.TemplatesCount/2 - 1
	} else {
		expectedCount = g.TemplatesCount / 2
	}
	if len(*templatesArr) != expectedCount {
		return fmt.Errorf("array length %d, expect %d", len(*templatesArr), expectedCount)
	}
	return nil
}
