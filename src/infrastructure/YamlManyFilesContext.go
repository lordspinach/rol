// Package infrastructure contains all implementations
package infrastructure

import (
	"fmt"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"reflect"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/domain"
	"strconv"
	"strings"
)

//YamlManyFilesContext implementation of interfaces.IDataContext which work with .yaml files
// and store each entity in own file
type YamlManyFilesContext[ItemIdType comparable, ItemType interfaces.IEntityModel[ItemIdType]] struct {
	DirPath          string
	inMemItems       map[ItemIdType]ItemType
	inMemItemsBackup map[ItemIdType]ItemType
}

//NewYamlManyFilesContext constructor for YamlManyFilesContext
func NewYamlManyFilesContext[ItemIdType comparable, ItemType interfaces.IEntityModel[ItemIdType]](diParams domain.GlobalDIParameters, dirPath string) (*YamlManyFilesContext[ItemIdType, ItemType], error) {
	err := os.MkdirAll(diParams.RootPath+dirPath, 0777)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to create directories")
	}
	context := &YamlManyFilesContext[ItemIdType, ItemType]{
		DirPath: diParams.RootPath + dirPath,
	}
	context.inMemItems = make(map[ItemIdType]ItemType)
	err = context.readFiles()
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to read files")
	}
	return context, nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) fromGenericToCommonType(id any) string {
	switch convertedID := id.(type) {
	case string:
		return convertedID
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", convertedID)
	case uuid.UUID:
		return convertedID.String()
	default:
		panic("type does not implemented")
	}
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) getLastID() (ItemIdType, error) {
	files, err := ioutil.ReadDir(c.DirPath)
	if err != nil {
		return *new(ItemIdType), errors.Internal.Wrap(err, "failed to read directory")
	}
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), "last.id.") {
			sep := strings.Split(file.Name(), ".")

			id, err := strconv.Atoi(sep[2])
			if err != nil {
				return *new(ItemIdType), errors.Internal.Wrap(err, "failed to parse string to int")
			}
			oldName := fmt.Sprintf(c.DirPath + file.Name())
			newName := fmt.Sprintf(c.DirPath+"last.id.%d", id+1)
			err = os.Rename(oldName, newName)
			if err != nil {
				return *new(ItemIdType), errors.Internal.Wrap(err, "failed to rename file")
			}
			return c.stringToItemIDType(fmt.Sprintf("%d", id))
		}
	}
	err = os.WriteFile(c.DirPath+"last.id.1", nil, 0644)
	if err != nil {
		return *new(ItemIdType), errors.Internal.Wrap(err, "failed to write file")
	}
	return c.stringToItemIDType("0")
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) newID() (ItemIdType, error) {
	t := *new(ItemIdType)
	var out any
	switch any(t).(type) {
	case string:
		out = uuid.New().String()
	case int:
		lastID, err := c.getLastID()
		if err != nil {
			return *new(ItemIdType), errors.Internal.Wrap(err, "failed to get last int id")
		}
		out = c.itemIDTypeToInt(lastID) + 1
	case uuid.UUID:
		out = uuid.New()
	}
	return out.(ItemIdType), nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) itemIDTypeToInt(id ItemIdType) int {
	switch any(id).(type) {
	case int:
		return any(id).(int)
	}
	panic("wrong generic type passed")
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) stringToItemIDType(id string) (ItemIdType, error) {
	t := *new(ItemIdType)
	var out any
	switch any(t).(type) {
	case string:
		out = uuid.New().String()
	case int:
		var err error
		out, err = strconv.Atoi(id)
		if err != nil {
			return *new(ItemIdType), errors.Internal.Wrap(err, "failed to parse string to int")
		}
	case uuid.UUID:
		out = uuid.New()
	default:
		panic("id type not implemented")
	}
	return out.(ItemIdType), nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) readFiles() error {
	files, err := ioutil.ReadDir(c.DirPath)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to read directory")
	}
	for _, file := range files {
		if !file.IsDir() && !strings.Contains(file.Name(), "last.id.") {
			item, err := ReadYamlFile[ItemType](c.DirPath + file.Name())
			if err != nil {
				return errors.Internal.Wrap(err, "failed to parse yaml file to struct")
			}
			nameWithExt := file.Name()
			extensionIndex := strings.LastIndex(nameWithExt, ".")
			name := nameWithExt[:extensionIndex]
			id, err := c.stringToItemIDType(name)
			if err != nil {
				return errors.Internal.Wrap(err, "failed to convert string to generic type")
			}
			c.inMemItems[id] = item
		}
	}
	return nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) setIDToItem(id ItemIdType, item *ItemType) {
	itemReflect := reflect.ValueOf(item).Elem()
	itemReflect.FieldByName("ID").Set(reflect.ValueOf(id))
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) restoreBackup() {
	restoreMap := make(map[ItemIdType]ItemType)
	for id, item := range c.inMemItemsBackup {
		restoreMap[id] = item
	}
	c.inMemItems = restoreMap
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) backupItems() {
	backupMap := make(map[ItemIdType]ItemType)
	for id, item := range c.inMemItems {
		backupMap[id] = item
	}
	c.inMemItemsBackup = backupMap
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) getFromMemory(id ItemIdType) (ItemType, error) {
	if !c.itemExist(id) {
		return *new(ItemType), errors.NotFound.New("item with given id was not found")
	}
	return c.inMemItems[id], nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) deleteFromMemory(id ItemIdType) {
	delete(c.inMemItems, id)
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) addToMemory(id ItemIdType, item ItemType) {
	c.inMemItems[id] = item
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) itemExist(id ItemIdType) bool {
	_, found := c.inMemItems[id]
	return found
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) saveToDisk(item ItemType) error {
	itemName := c.fromGenericToCommonType(item.GetID())
	err := SaveYamlFile(item, c.DirPath+itemName+".yaml")
	if err != nil {
		return errors.Internal.Wrap(err, "failed to save yaml file")
	}
	return nil
}

func (c *YamlManyFilesContext[ItemIdType, ItemType]) deleteFromDisk(id ItemIdType) error {
	itemName := c.fromGenericToCommonType(id)
	err := os.Remove(c.DirPath + itemName + ".yaml")
	if err != nil {
		return errors.Internal.Wrap(err, "failed to remove file from disk")
	}
	return nil
}

//Get all items
//
//Return
//	map[ItemIdType]ItemType - items map
func (c *YamlManyFilesContext[ItemIdType, ItemType]) Get() (map[ItemIdType]ItemType, error) {
	return c.inMemItems, nil
}

//GetByID get item by id
//
//Params
//	id - item id
//Return
//	ItemType - found item
//	error - if an error occurs, otherwise nil
func (c *YamlManyFilesContext[ItemIdType, ItemType]) GetByID(id ItemIdType) (ItemType, error) {
	return c.getFromMemory(id)
}

//Add new item
//
//Params
//	item - item to add
//Return
//	ItemType - added item
//	error - if an error occurs, otherwise nil
func (c *YamlManyFilesContext[ItemIdType, ItemType]) Add(item ItemType) (ItemType, error) {
	if c.itemExist(item.GetID()) {
		return *new(ItemType), errors.AlreadyExist.New("item with given id already exist")
	}
	id, err := c.newID()
	if err != nil {
		return *new(ItemType), errors.Internal.Wrap(err, "failed to create new id")
	}
	c.backupItems()
	c.setIDToItem(id, &item)
	c.addToMemory(id, item)
	if err = c.saveToDisk(item); err != nil {
		c.restoreBackup()
		return *new(ItemType), errors.Internal.Wrap(err, "failed to write file to disk")
	}
	return item, nil
}

//Update item by id
//
//Params
//	id - item id
//	item - item to update
//Return
//	error - if an error occurs, otherwise nil
func (c *YamlManyFilesContext[ItemIdType, ItemType]) Update(id ItemIdType, item ItemType) error {
	if !c.itemExist(id) {
		return errors.NotFound.New("item with given id was not found")
	}
	c.backupItems()
	c.addToMemory(id, item)
	if err := c.saveToDisk(item); err != nil {
		c.restoreBackup()
		return errors.Internal.Wrap(err, "failed to write file to disk")
	}
	return nil

}

//Delete item by id
//
//Params
//	id - item id
//Return
//	error - if an error occurs, otherwise nil
func (c *YamlManyFilesContext[ItemIdType, ItemType]) Delete(id ItemIdType) error {
	if !c.itemExist(id) {
		return errors.NotFound.New("item with given id was not found")
	}
	c.backupItems()
	if err := c.deleteFromDisk(id); err != nil {
		c.restoreBackup()
		return errors.Internal.Wrap(err, "failed to delete file from disk")
	}
	c.deleteFromMemory(id)
	return nil
}
