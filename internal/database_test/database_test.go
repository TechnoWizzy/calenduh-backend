package database_test

// import (
// 	"calenduh-backend/internal/database"
// 	"calenduh-backend/internal/sqlc"
// 	"context"
// 	"testing"
// )

// func TestCreateEvent(t *testing.T) {
//     instance, err := database.New("POSTGRESQL_URL")
//     if err != nil {
//         t.Fatalf("failed to connect to database: %v", err)
//     }

//     location := "Test Location"
//     description := "Test Description"
//     notification := "Test Notification"
//     frequency_ := int32(1)
//     frequency := &frequency_
//     priority := int32(1)
//     event := sqlc.Event{
//         EventID:    "event123",
//         CalendarID: "calendar123",
//         Name:       "Test Event",
//         Location:   &location,
//         Description: &description,
//         Notification: &notification,
//         Frequency:  frequency,
//         Priority:   &priority,
//     }
       
//     err = instance.CreateEvent(context.Background(), event)
//     if err != nil {
//         t.Fatalf("Error creating event: %v", err)
//     }
// }
