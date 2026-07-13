package documents

import (
	"time"

	"github.com/google/uuid"
)

type DocumentCommentResponse struct {
	ID           uuid.UUID                 `json:"id"`
	DocumentID   uuid.UUID                 `json:"documentId"`
	UserID       uuid.UUID                 `json:"userId"`
	UserFullName string                    `json:"userFullName"`
	ParentID     *uuid.UUID                `json:"parentId"`
	Content      string                    `json:"content"`
	IsEdited     bool                      `json:"isEdited"`
	CreatedAt    time.Time                 `json:"createdAt"`
	UpdatedAt    time.Time                 `json:"updatedAt"`
	Replies      []DocumentCommentResponse `json:"replies"`
}

type CreateDocumentCommentRequest struct {
	Content  string     `json:"content"`
	ParentID *uuid.UUID `json:"parentId"`
}

type UpdateDocumentCommentRequest struct {
	Content string `json:"content"`
}

func toDocumentCommentResponse(row CommentListRow) DocumentCommentResponse {
	return DocumentCommentResponse{
		ID:           row.ID,
		DocumentID:   row.DocumentID,
		UserID:       row.UserID,
		UserFullName: row.UserFullName,
		ParentID:     row.ParentID,
		Content:      row.Content,
		IsEdited:     row.IsEdited,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		Replies:      []DocumentCommentResponse{},
	}
}

func buildCommentThread(rows []CommentListRow) []DocumentCommentResponse {
	byID := make(map[uuid.UUID]DocumentCommentResponse, len(rows))
	children := make(map[uuid.UUID][]DocumentCommentResponse)

	for _, row := range rows {
		byID[row.ID] = toDocumentCommentResponse(row)
	}

	var roots []DocumentCommentResponse
	for _, row := range rows {
		node := byID[row.ID]
		if row.ParentID == nil {
			roots = append(roots, node)
			continue
		}
		if _, ok := byID[*row.ParentID]; ok {
			children[*row.ParentID] = append(children[*row.ParentID], node)
			continue
		}
		roots = append(roots, node)
	}

	var attach func(DocumentCommentResponse) DocumentCommentResponse
	attach = func(node DocumentCommentResponse) DocumentCommentResponse {
		if replies, ok := children[node.ID]; ok {
			node.Replies = make([]DocumentCommentResponse, 0, len(replies))
			for _, reply := range replies {
				node.Replies = append(node.Replies, attach(reply))
			}
		}
		return node
	}

	result := make([]DocumentCommentResponse, 0, len(roots))
	for _, root := range roots {
		result = append(result, attach(root))
	}
	return result
}
