package controllers

import (
	"calenduh-backend/internal/database"
	"calenduh-backend/internal/sqlc"
	"github.com/gin-gonic/gin"
	"net/http"
)

type MergeAccount struct {
	UserId 		string		`json:"user_id"`
	Calendars   []string	`json:"calendars"`
	Events      []string	`json:"events"`
}

func Merge(c *gin.Context) {
	var input MergeAccount
	if err := c.BindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.Transaction(c, func(queries *sqlc.Queries) error {
		// calendars first
		for _, calendarId := range input.Calendars {
			if err := queries.TransferCalendarOwnership(c, sqlc.TransferCalendarOwnershipParams{
				CalendarID: calendarId,
				UserID: &input.UserId,
			}); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return err
			}
		}
		// note: events are stored in the calendar, so no need to transfer them
		c.JSON(http.StatusOK/*more descriptive later?*/, gin.H{"message": "Merge successful"})
		return nil
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}