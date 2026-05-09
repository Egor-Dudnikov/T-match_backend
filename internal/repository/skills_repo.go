package repository

import (
	"T-match_backend/internal/models"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
)

func (r *Repository) AddInternSkills(ctx context.Context, skills []int, userID int) error {
	id, err := r.GetInternId(ctx, userID)
	if err != nil {
		return err
	}
	if len(skills) == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString("INSERT INTO intern_skills (intern_id, skill_id) VALUES ")
	delimiter := false
	cnt := 1

	values := []interface{}{id}

	for _, value := range skills {
		cnt++
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("($1, $")
		query.WriteString(strconv.Itoa(cnt))
		query.WriteString(")")
		delimiter = true
		values = append(values, value)
	}

	skills = append(skills)
	_, err = r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetInternSkills(ctx context.Context, userID int) ([]models.Skill, error) {
	id, err := r.GetInternId(ctx, userID)
	res := []models.Skill{}
	skillIDs := []int{}
	if err != nil {
		return res, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT skill_id FROM intern_skills WHERE intern_id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, nil
		}
		return res, nil
	}
	defer rows.Close()

	for rows.Next() {
		var skillID int
		rows.Scan(&skillID)
		skillIDs = append(skillIDs, skillID)
	}

	res, err = r.GetNameSkills(ctx, skillIDs)
	return res, err
}

func (r *Repository) DeleteInternSkills(ctx context.Context, skills []int, userID int) error {
	if len(skills) == 0 {
		return nil
	}

	id, err := r.GetInternId(ctx, userID)
	if err != nil {
		return err
	}

	var query strings.Builder
	query.WriteString("DELETE FROM intern_skills WHERE intern_id = $1 AND skill_id IN (")
	delimiter := false
	values := []interface{}{id}
	cnt := 1

	for _, value := range skills {
		cnt++
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("$")
		query.WriteString(strconv.Itoa(cnt))

		values = append(values, value)

		delimiter = true
	}
	query.WriteString(")")

	_, err = r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetInternId(ctx context.Context, UserID int) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM interns WHERE user_id = $1`, UserID).Scan(&id)
	return id, err
}

func (r *Repository) AddInternshipSkills(ctx context.Context, skills []int, internshipID int) error {
	if len(skills) == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString("INSERT INTO internship_skills (internship_id, skill_id) VALUES ")
	delimiter := false
	cnt := 1

	values := []interface{}{internshipID}

	for _, value := range skills {
		cnt++
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("($1, $")
		query.WriteString(strconv.Itoa(cnt))
		query.WriteString(")")
		delimiter = true
		values = append(values, value)
	}

	skills = append(skills)
	_, err := r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetInternshipSkills(ctx context.Context, internshipID int) ([]models.Skill, error) {
	res := []models.Skill{}
	skillIDs := []int{}
	rows, err := r.db.QueryContext(ctx, `SELECT skill_id FROM internship_skills WHERE internship_id = $1`, internshipID)
	defer rows.Close()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, nil
		}
		return res, nil
	}

	for rows.Next() {
		var skillID int
		rows.Scan(&skillID)
		skillIDs = append(skillIDs, skillID)
	}

	res, err = r.GetNameSkills(ctx, skillIDs)
	return res, err
}

func (r *Repository) DeleteInternshipSkills(ctx context.Context, skills []int, internshipID int) error {
	if len(skills) == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString("DELETE FROM internship_skills WHERE internship_id = $1 AND skill_id IN (")
	delimiter := false
	values := []interface{}{internshipID}
	cnt := 1

	for _, value := range skills {
		cnt++
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("$")
		query.WriteString(strconv.Itoa(cnt))

		values = append(values, value)

		delimiter = true
	}
	query.WriteString(")")

	_, err := r.db.ExecContext(ctx, query.String(), values...)
	return err
}

func (r *Repository) GetNameSkills(ctx context.Context, skillsID []int) ([]models.Skill, error) {
	res := []models.Skill{}
	if len(skillsID) == 0 {
		return res, nil
	}

	var query strings.Builder
	query.WriteString("SELECT name FROM skills WHERE id IN (")
	delimiter := false
	values := []interface{}{}
	for i, value := range skillsID {
		values = append(values, value)
		res = append(res, models.Skill{Id: value})
		if delimiter {
			query.WriteString(", ")
		}
		query.WriteString("$")
		query.WriteString(strconv.Itoa(i + 1))
		delimiter = true
	}
	query.WriteString(")")

	rows, err := r.db.QueryContext(ctx, query.String(), values...)
	if err != nil {
		return res, err
	}
	defer rows.Close()

	ptr := 0
	for rows.Next() {
		rows.Scan(&res[ptr].Name)
		ptr++
	}

	return res, nil
}
