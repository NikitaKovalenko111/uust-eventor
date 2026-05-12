package models

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

type User struct {
	ID            uint64         `json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	PasswordHash  string         `json:"-"`
	Role          string         `json:"role"` // 'user' или 'moderator'
	About         sql.NullString `json:"about,omitempty"`
	City          string         `json:"city"`
	Faculty       sql.NullString `json:"faculty,omitempty"`
	Course        sql.NullString `json:"course,omitempty"`
	AvatarImageID sql.NullString `json:"avatar_image_id,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// TableName возвращает имя таблицы для ORM (если используется)
func (User) TableName() string {
	return "users"
}

// Validate проверяет все бизнес-правила перед созданием/обновлением пользователя
// Возвращает первую обнаруженную ошибку (для простоты; можно собирать все через []error)
func (u *User) Validate(isCreate bool) error {
	// =====================================================
	// 1. ОБЯЗАТЕЛЬНЫЕ ПОЛЯ (строки)
	// =====================================================

	// Name: обязателен, 2-255 символов, только буквы/пробелы/дефисы
	if err := validateName(u.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}

	// Email: обязателен, валидный формат, ≤255 символов
	if err := validateEmail(u.Email); err != nil {
		return fmt.Errorf("email: %w", err)
	}

	// City: обязателен, 2-128 символов
	if err := validateCity(u.City); err != nil {
		return fmt.Errorf("city: %w", err)
	}

	// Role: обязателен, только 'user' или 'moderator'
	if err := validateRole(u.Role); err != nil {
		return fmt.Errorf("role: %w", err)
	}

	// PasswordHash: обязателен только при создании (не при обновлении профиля)
	if isCreate {
		if err := validatePasswordHash(u.PasswordHash); err != nil {
			return fmt.Errorf("password: %w", err)
		}
	}

	// =====================================================
	// 2. ОПЦИОНАЛЬНЫЕ ПОЛЯ (sql.NullString)
	// =====================================================

	// About: если указано — не пустое, ≤1000 символов (разумный лимит для TEXT)
	if u.About.Valid {
		if err := validateOptionalText(u.About.String, 1000, "about"); err != nil {
			return fmt.Errorf("about: %w", err)
		}
	}

	// Faculty: если указано — не пустое, ≤255 символов, только буквы/пробелы/дефисы/точки
	if u.Faculty.Valid {
		if err := validateFaculty(u.Faculty.String); err != nil {
			return fmt.Errorf("faculty: %w", err)
		}
	}

	// Course: если указано — валидный формат (1-6, или "1к", "2бакалавр"), ≤10 символов
	if u.Course.Valid {
		if err := validateCourse(u.Course.String); err != nil {
			return fmt.Errorf("course: %w", err)
		}
	}

	// AvatarImageID: если указано — валидный UUID или непустая строка ≤128 символов
	if u.AvatarImageID.Valid {
		if err := validateAvatarID(u.AvatarImageID.String); err != nil {
			return fmt.Errorf("avatar_image_id: %w", err)
		}
	}

	// =====================================================
	// 3. БИЗНЕС-ПРАВИЛА (кросс-поля)
	// =====================================================

	// Если роль 'moderator' — faculty и course должны быть заполнены (пример правила)
	// (Раскомментируйте, если нужно)
	/*
		if u.Role == "moderator" {
			if !u.Faculty.Valid || u.Faculty.String == "" {
				return errors.New("moderators must specify faculty")
			}
			if !u.Course.Valid || u.Course.String == "" {
				return errors.New("moderators must specify course")
			}
		}
	*/

	// Курс должен соответствовать факультету (если нужна сложная логика — вынести в сервис)

	// =====================================================
	// 4. ТЕХНИЧЕСКИЕ ПОЛЯ (не валидируем, заполняются БД)
	// =====================================================
	// CreatedAt, UpdatedAt, ID — игнорируем, они управляются БД

	return nil
}

// =====================================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ВАЛИДАЦИИ
// =====================================================

// validateName проверяет имя пользователя
func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("is required")
	}
	if len(name) < 2 {
		return errors.New("must be at least 2 characters")
	}
	if len(name) > 255 {
		return errors.New("must not exceed 255 characters")
	}
	// Разрешаем: буквы (в т.ч. кириллица), пробелы, дефисы, апострофы
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return errors.New("contains invalid characters (only letters, spaces, hyphens, apostrophes allowed)")
		}
	}
	return nil
}

// validateEmail проверяет email
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return errors.New("is required")
	}
	if len(email) > 255 {
		return errors.New("must not exceed 255 characters")
	}
	// Простая проверка формата (для продакшена используйте библиотеку like govalidator)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("has invalid format")
	}
	// Дополнительно: нижний регистр для консистентности (опционально)
	// email = strings.ToLower(email)
	return nil
}

// validateCity проверяет город
func validateCity(city string) error {
	city = strings.TrimSpace(city)
	if city == "" {
		return errors.New("is required")
	}
	if len(city) < 2 {
		return errors.New("must be at least 2 characters")
	}
	if len(city) > 128 {
		return errors.New("must not exceed 128 characters")
	}
	// Разрешаем буквы, пробелы, дефисы, точки (для "Санкт-Петербург", "New York")
	for _, r := range city {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '.' {
			return errors.New("contains invalid characters")
		}
	}
	return nil
}

// validateRole проверяет роль
func validateRole(role string) error {
	role = strings.TrimSpace(role)
	if role == "" {
		return errors.New("is required")
	}
	if role != "user" && role != "moderator" {
		return fmt.Errorf("invalid value '%s' (allowed: 'user', 'moderator')", role)
	}
	return nil
}

// validatePasswordHash проверяет хеш пароля
// Обычно валидация пароля (длина, сложность) делается ДО хеширования, на уровне сервиса
func validatePasswordHash(hash string) error {
	if hash == "" {
		return errors.New("is required")
	}
	// bcrypt-хеши начинаются с $2a$, $2b$ и имеют длину ~60 символов
	// Проверяем минимальную длину для защиты от ошибок
	if len(hash) < 20 {
		return errors.New("appears to be invalid (too short for bcrypt/argon2 hash)")
	}
	return nil
}

// validateOptionalText проверяет опциональное текстовое поле
func validateOptionalText(value string, maxLength int, fieldName string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil // пустое значение допустимо для опциональных полей
	}
	if len(value) > maxLength {
		return fmt.Errorf("must not exceed %d characters", maxLength)
	}
	return nil
}

// validateFaculty проверяет факультет
func validateFaculty(faculty string) error {
	faculty = strings.TrimSpace(faculty)
	if faculty == "" {
		return errors.New("cannot be empty string (use NULL if not specified)")
	}
	if len(faculty) > 255 {
		return errors.New("must not exceed 255 characters")
	}
	// Разрешаем буквы, пробелы, дефисы, точки, скобки (для "ФКН (ВШЭ)", "Мех-мат МГУ")
	for _, r := range faculty {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' && r != '-' && r != '.' && r != '(' && r != ')' {
			return errors.New("contains invalid characters")
		}
	}
	return nil
}

// validateCourse проверяет курс
func validateCourse(course string) error {
	course = strings.TrimSpace(course)
	if course == "" {
		return errors.New("cannot be empty string (use NULL if not specified)")
	}
	if len(course) > 10 {
		return errors.New("must not exceed 10 characters")
	}
	// Разрешаем форматы: "1", "2к", "3бакалавр", "4магистр", "1-й"
	// Простая проверка: начинается с цифры 1-6
	courseRegex := regexp.MustCompile(`^[1-6][0-9а-яА-Яa-zA-Z\-]*$`)
	if !courseRegex.MatchString(course) {
		return errors.New("has invalid format (expected: 1-6, optionally followed by text like 'к', 'бакалавр')")
	}
	return nil
}

// validateAvatarID проверяет ID аватара
func validateAvatarID(avatarID string) error {
	avatarID = strings.TrimSpace(avatarID)
	if avatarID == "" {
		return errors.New("cannot be empty string (use NULL if not specified)")
	}
	if len(avatarID) > 128 {
		return errors.New("must not exceed 128 characters")
	}
	// Если используете UUID — раскомментируйте проверку:
	/*
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(strings.ToLower(avatarID)) {
			return errors.New("must be a valid UUID")
		}
	*/
	return nil
}
