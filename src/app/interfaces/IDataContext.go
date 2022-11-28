package interfaces

//IDataContext is an interface for custom data context
type IDataContext[ItemIdType comparable, ItemType IEntityModel[ItemIdType]] interface {
	//Add new item
	//
	//Params
	//	item - item to add
	//Return
	//	ItemType - added item
	//	error - if an error occurs, otherwise nil
	Add(item ItemType) (ItemType, error)
	//Update item by id
	//
	//Params
	//	id - item id
	//	item - item to update
	//Return
	//	error - if an error occurs, otherwise nil
	Update(id ItemIdType, item ItemType) error
	//Delete item by id
	//
	//Params
	//	id - item id
	//Return
	//	error - if an error occurs, otherwise nil
	Delete(id ItemIdType) error
	//Get all items
	//
	//Return
	//	map[ItemIdType]ItemType - items map
	Get() (map[ItemIdType]ItemType, error)
	//GetByID get item by id
	//
	//Params
	//	id - item id
	//Return
	//	ItemType - found item
	//	error - if an error occurs, otherwise nil
	GetByID(id ItemIdType) (ItemType, error)
}
