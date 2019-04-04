package main

import (
	"log"
	"time"

	models "./db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

type LastTenQuestions struct {
	ID       []uint
	Question []string
}

type LastTenAnswers struct {
	ID         []uint
	Answer     []string
	QuestionID []int
}

//InitializeDB sets up the mySQL connection
func InitializeDB() (db *gorm.DB) {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Printf("Successfully connected to qBot Database.")

	db.DB().SetConnMaxLifetime(time.Second * 100)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(200)

	//db.DropTableIfExists(models.User{}, models.Question{}, models.Answer{}) //Temp

	if err := db.AutoMigrate(models.User{}, models.Question{}, models.Answer{}).Error; err != nil {
		log.Fatalf("Unable to migrate database. \nReason: %v", err)
	}
	log.Printf("Migrating Database...")
	return db
}

func (qb *qBot) CreateNewDBRecord(record interface{}) error {
	if qb.DB.NewRecord(record) != true {
		log.Printf("The value's primary key is not blank")
	}
	if err := qb.DB.Create(record).Error; err != nil {
		log.Printf("Unable to create new Database record")
		return err
	}
	log.Printf("A new Database Record were successfully added.")
	return nil
}

func (qb *qBot) UserExistInDB(newUserRecord models.User) bool {
	var count int64
	//Count DB entries matching the Slack User ID
	if err := qb.DB.Where("slack_user = ?", newUserRecord.SlackUser).First(&newUserRecord).Count(&count); err != nil {
		if count == 0 { //Avoid duplicate User entries in the DB.
			return false
		}
	}
	return true
}

func (qb *qBot) LastTenQuestions(ltq *LastTenQuestions) (LastTenQuestions, error) {
	tenQuestions, _ := qb.DB.Model(&models.Question{}).Last(&[]models.Question{}).Limit(10).Rows()
	for tenQuestions.Next() {
		q := new(models.Question)
		err := qb.DB.ScanRows(tenQuestions, q)
		if err != nil {
			log.Printf("Unable to parse SQL query into a crunchable dataformat. \nReason: %v", err)
		}
		log.Printf("Question %v: %s\n", q.ID, q.Question)

		ltq.ID = append(ltq.ID, q.ID)
		ltq.Question = append(ltq.Question, q.Question)
	}
	return *ltq, nil
}

func (qb *qBot) LastTenAnswers(lta *LastTenAnswers) (LastTenAnswers, error) {
	tenAnswers, _ := qb.DB.Model(&[]*models.Question{}).Related(&models.Answer{}, "Answers").Last(&[]models.Answer{}).Limit(10).Rows()
	for tenAnswers.Next() {
		a := new(models.Answer)
		err := qb.DB.ScanRows(tenAnswers, a)
		if err != nil {
			log.Printf("Unable to parse SQL query into a crunchable dataformat. \nReason: %v", err)
		}
		log.Printf("Answer %v: %s, %v\n", a.ID, a.Answer, a.QuestionID)
		lta.ID = append(lta.ID, a.ID)
		lta.Answer = append(lta.Answer, a.Answer)
		lta.QuestionID = append(lta.QuestionID, a.QuestionID)
	}
	return *lta, nil
}
