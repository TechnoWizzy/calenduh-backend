package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"net/http"
)

func GetAllGroups(c *gin.Context) {
	groups, err := database.Db.Queries.GetAllGroups(c)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.JSON(http.StatusOK, make([]sqlc.Group, 0))
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, groups)
}

func GetMyGroups(c *gin.Context) {
	groups := *ParseGroups(c)
	c.PureJSON(http.StatusOK, groups)
}

func GetGroup(c *gin.Context) {
	groups := *ParseGroups(c)
	groupId := c.Param("group_id")
	if groupId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	for _, group := range groups {
		if group.GroupID == groupId {
			c.JSON(http.StatusOK, group)
			return
		}
	}

	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found or not permissible"})
}

func CreateGroup(c *gin.Context) {
	user := *ParseUser(c)
	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		var input sqlc.CreateGroupParams
		if err := c.BindJSON(&input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return err
		}

		input.GroupID = gonanoid.Must()

		group, err := queries.CreateGroup(c, input)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}

		params := sqlc.CreateGroupMemberParams{
			GroupID: group.GroupID,
			UserID:  user.UserID,
		}
		if err = queries.CreateGroupMember(c, params); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return err
		}

		c.JSON(http.StatusCreated, group)
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func JoinGroup(c *gin.Context) {
	user := *ParseUser(c)
	inviteCode := c.Param("invite_code")

	if inviteCode == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invite_code is required"})
		return
	}

	group, err := database.Db.Queries.GetGroupByInviteCode(c, inviteCode)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if err = database.Db.Queries.CreateGroupMember(c, sqlc.CreateGroupMemberParams{
		UserID:  user.UserID,
		GroupID: group.GroupID,
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

func LeaveGroup(c *gin.Context) {
	user := *ParseUser(c)
	groups := *ParseGroups(c)
	groupId := c.Param("group_id")

	if groupId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		for _, group := range groups {
			members, err := queries.GetGroupMembers(c, group.GroupID)
			if err != nil {
				return err
			}
			if groupId == group.GroupID {
				if err := queries.DeleteGroupMember(c, sqlc.DeleteGroupMemberParams{
					UserID:  user.UserID,
					GroupID: group.GroupID,
				}); err != nil {
					return err
				}

				if len(members) == 1 {
					if err := queries.DeleteGroup(c, group.GroupID); err != nil {
						return err
					}
				}

				c.Status(http.StatusOK)
				return nil
			}
		}
		return errors.New("cannot leave group you are not in")
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return
}

func UpdateGroup(c *gin.Context) {
	groups := *ParseGroups(c)
	var input sqlc.UpdateGroupParams
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	input.GroupID = c.Param("group_id")
	if input.GroupID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
	}
	if !CanEditGroup(input.GroupID, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	group, err := database.Db.Queries.UpdateGroup(c, input)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, group)
}

func DeleteGroup(c *gin.Context) {
	groups := *ParseGroups(c)
	groupId := c.Param("group_id")
	if groupId == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
	}

	if !CanEditGroup(groupId, groups) {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err := database.Db.Queries.DeleteGroup(c, groupId)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "group not found"})
		default:
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func ParseGroups(c *gin.Context) *[]sqlc.Group {
	v, found := c.Get("groups")
	if !found {
		panic(errors.New("groups not found"))
	}
	groups, ok := v.(*[]sqlc.Group)
	if !ok {
		panic(errors.New("groups type assertion failed"))
	}

	return groups
}

func CanEditGroup(groupId string, groups []sqlc.Group) bool {
	for _, group := range groups {
		if group.GroupID == groupId {
			return true
		}
	}

	return false
}
