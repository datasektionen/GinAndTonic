package models

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name          string         `json:"name"`
	Email         string         `json:"email" gorm:"unique"`
	Events        []Event        `gorm:"foreignKey:TeamID" json:"events"`
	Users         []User         `gorm:"many2many:team_users;" json:"users"`
	TeamUserRoles []TeamUserRole `gorm:"foreignKey:TeamID" json:"team_user_roles"`
	BankingDetail BankingDetail  `json:"banking_detail" gorm:"foreignKey:TeamID"`
}

func (Team) TableName() string {
	return "teams"
}

func CreateTeamUniqueIndex(db *gorm.DB) error {
	// Adjust the line below to your actual table name, if it's different
	const tableName = "teams"

	// Drop the old unique constraint if it exists
	if err := db.Exec(`ALTER TABLE ` + tableName + ` DROP CONSTRAINT IF EXISTS idx_teams_name_unique;`).Error; err != nil {
		return err
	}

	// Create a new unique index
	if err := db.Exec(`CREATE UNIQUE INDEX idx_teams_name_unique ON ` + tableName + ` (name) WHERE deleted_at IS NULL;`).Error; err != nil {
		fmt.Println("Error creating unique index on team.name:", err)
	}

	return nil
}

func (o Team) ValidateName() error {
	if len(o.Name) < 3 {
		return errors.New("name must be at least 3 characters long")
	}
	if len(o.Name) > 50 {
		return errors.New("name must not be longer than 50 characters")
	}
	if match, _ := regexp.MatchString(`^[\p{L}\p{N} -]*$`, o.Name); !match {
		return errors.New("name must not contain special characters")
	}
	return nil
}

func (o Team) Validate() error {
	if err := o.ValidateName(); err != nil {
		return err
	}
	return nil
}

func GetTeamByNameIfExist(db *gorm.DB, name string) (Team, error) {
	var team Team
	err := db.Preload("Events").Where("name = ?", name).First(&team).Error
	return team, err
}

func GetTeamByIDIfExist(db *gorm.DB, id uint) (Team, error) {
	var team Team
	err := db.Preload("Events").Where("id = ?", id).First(&team).Error
	return team, err
}

func (o Team) GetUsers(db *gorm.DB) ([]User, error) {
	var users []User

	// Preload the Users and TeamUserRoles for the team
	if err := db.Model(&o).Preload("TeamUserRoles").Association("Users").Find(&users); err != nil {
		return nil, err
	}

	// Remove the user.TeamUserRoles that are not for this team
	for i, user := range users {
		var roles []TeamUserRole
		for _, role := range user.TeamUserRoles {
			if role.TeamID == o.ID {
				roles = append(roles, role)
			}
		}
		users[i].TeamUserRoles = roles
	}

	return users, nil
}

func GetTeamOwners(db *gorm.DB, team Team) (users []User, err error) {
	for _, user := range team.Users {
		for _, role := range user.TeamUserRoles {

			if role.TeamRoleName == string(TeamOwner) && role.TeamID == team.ID {
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func GetAllTeamEvents(db *gorm.DB, orgId uint) (events []Event, err error) {
	err = db.Where("team_id = ?", orgId).Find(&events).Error
	return
}
