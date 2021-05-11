package infrastructure

import (
	"database/sql"
	"goinv/domain"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func (m *MySQLUserRepository) Find(ID uuid.UUID) (*domain.Task, error) {
	row := m.db.QueryRow("SELECT * FROM tasks WHERE ?", ID.String())
	task := &domain.Task{}
	err := row.Scan(task.ID, task.OriginURL, task.DownloadURL, task.Status)
	if err != nil {
		return nil, err
	}
	return task, nil
}
func (m *MySQLUserRepository) Save(task *domain.Task) error {
	insert, err := m.db.Query("INSERT IGNORE tasks(id, origin_url, download_url, status) VALUES (?,?,?)", task.ID, task.OriginURL, task.DownloadURL, task.Status)
	if err != nil {
		return err
	}
	insert.Close()
	return nil
}
func (m *MySQLUserRepository) FindAll() (domain.Tasks, error) {
	rows, err := m.db.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make(domain.Tasks, 0)
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(task.ID, task.OriginURL, task.DownloadURL, task.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}
