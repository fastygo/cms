package revisions

import (
	"time"

	"github.com/fastygo/cms/internal/domain/content"
)

type ID string

type Revision struct {
	ID        ID
	EntryID   content.ID
	Snapshot  content.Entry
	AuthorID  string
	Reason    string
	CreatedAt time.Time
}
