package db

import (
	"context"
	"errors"

	"github.com/cockroachdb/cockroach-go/crdb/crdbgorm"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func InitDB(debug bool, user string) (err error) {
	var addr = "postgresql://" + user +
		"@free-tier.gcp-us-central1.cockroachlabs.cloud:26257/defaultdb" +
		"?sslmode=verify-full" +
		"&sslrootcert=/home/aadi/root.crt" +
		"&options=--cluster%3Dclear-weasel-3066"

	db, err = gorm.Open("postgres", addr)

	db.LogMode(debug)
	db.AutoMigrate(&Workspace{}, &Backlink{})

	return
}

func DeinitDB() error {
	return db.Close()
}

type Workspace struct {
	gorm.Model

	SlackTeam string
	Backlinks []Backlink `gorm:"foreignKey:WorkspaceID"`
}

type Backlink struct {
	gorm.Model

	LinkName string
	NotionID string

	WorkspaceID uint
}

func GetWorkspaceInfo(teamName string) (info Workspace) {
	db.Where(&Workspace{SlackTeam: teamName}, "slackteam").Take(&info)

	backlinks := []Backlink{}
	db.Where(&Backlink{WorkspaceID: info.ID}, "workspaceid").Find(&backlinks)
	info.Backlinks = backlinks

	return
}

func GetNotionID(teamName string, backlinkName string) (string, error) {
	workspace := GetWorkspaceInfo(teamName)
	for _, backlink := range workspace.Backlinks {
		if backlink.LinkName == backlinkName {
			return backlink.NotionID, nil
		}
	}
	return "", errors.New("cannot find backlink")
}

func AddWorkspace(teamName string) error {
	return crdbgorm.ExecuteTx(context.Background(), db, nil,
		func(tx *gorm.DB) error {
			return db.Create(&Workspace{SlackTeam: teamName, Backlinks: []Backlink{}}).Error
		},
	)
}

func AddBacklinkToWorkspace(teamName string, backlink Backlink) error {
	return crdbgorm.ExecuteTx(context.Background(), db, nil,
		func(tx *gorm.DB) error {
			workspace := GetWorkspaceInfo(teamName)
			workspace.Backlinks = append(workspace.Backlinks, backlink)
			return db.Save(&workspace).Error
		},
	)
}

func BacklinkExists(teamName string, backlinkName string) bool {
	workspace := GetWorkspaceInfo(teamName)
	for _, backlink := range workspace.Backlinks {
		if backlink.LinkName == backlinkName {
			return true
		}
	}
	return false
}

func DropAllTables() {
	db.DropTableIfExists(&Workspace{})
	db.DropTableIfExists(&Backlink{})
}
