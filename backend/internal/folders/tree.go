package folders

import "github.com/google/uuid"

func indexFoldersByID(items []Folder) map[uuid.UUID]Folder {
	index := make(map[uuid.UUID]Folder, len(items))
	for _, item := range items {
		index[item.ID] = item
	}
	return index
}

func childrenByParent(items []Folder) map[uuid.UUID][]uuid.UUID {
	children := make(map[uuid.UUID][]uuid.UUID)
	for _, item := range items {
		if item.ParentID == nil {
			continue
		}
		parentID := *item.ParentID
		children[parentID] = append(children[parentID], item.ID)
	}
	return children
}

func CollectDescendantIDs(items []Folder, rootID uuid.UUID) []uuid.UUID {
	children := childrenByParent(items)
	result := make([]uuid.UUID, 0)
	queue := append([]uuid.UUID{}, children[rootID]...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		queue = append(queue, children[current]...)
	}
	return result
}

func IsDescendant(items []Folder, ancestorID, nodeID uuid.UUID) bool {
	if ancestorID == nodeID {
		return true
	}
	byID := indexFoldersByID(items)
	current, ok := byID[nodeID]
	if !ok {
		return false
	}
	for current.ParentID != nil {
		if *current.ParentID == ancestorID {
			return true
		}
		parent, exists := byID[*current.ParentID]
		if !exists {
			return false
		}
		current = parent
	}
	return false
}

func BuildBreadcrumb(items []Folder, folderID uuid.UUID) []Folder {
	byID := indexFoldersByID(items)
	path := make([]Folder, 0)
	current, ok := byID[folderID]
	if !ok {
		return path
	}
	for {
		path = append([]Folder{current}, path...)
		if current.ParentID == nil {
			break
		}
		parent, exists := byID[*current.ParentID]
		if !exists {
			break
		}
		current = parent
	}
	return path
}

func CollectFolderAndDescendantIDs(items []Folder, rootID uuid.UUID) []uuid.UUID {
	descendants := CollectDescendantIDs(items, rootID)
	return append([]uuid.UUID{rootID}, descendants...)
}
