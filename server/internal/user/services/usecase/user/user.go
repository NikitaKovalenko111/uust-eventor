package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	erors "eventor/internal/user/domain/errors"
	"eventor/internal/user/domain/models"
	user_storage "eventor/internal/user/storage/repositories/user"
	user_dto "eventor/internal/user/transport/http/dto/user"

	// ваш интерфейс репозитория
	"golang.org/x/crypto/bcrypt"
)

// UserService бизнес-логика работы с пользователями
type UserService struct {
	repo *user_storage.UserRepo
}

// Init создаёт новый сервис
func Init(repo *user_storage.UserRepo) *UserService {
	return &UserService{repo: repo}
}

// =====================================================
// CREATE
// =====================================================

func (s *UserService) Create(ctx context.Context, req *user_dto.CreateUserRequest) (*models.User, error) {
	// 1. Нормализация
	req = s.normalizeCreateRequest(req)

	// 2. Валидация
	if err := s.validateCreateRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %w", erors.ErrValidation, err)
	}

	// 3. Конвертация в доменную модель
	user := req.ToDomain()

	// 4. Хеширование пароля
	hashed, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", erors.ErrHashing, err)
	}
	user.PasswordHash = hashed

	// 5. Проверка уникальности email (pre-check)
	if err := s.checkEmailUnique(ctx, user.Email, 0); err != nil {
		return nil, err
	}

	// 6. Сохранение
	if err := s.repo.Create(ctx, user); err != nil {
		if isUniqueViolation(err, "email") {
			return nil, fmt.Errorf("%w: %s", erors.ErrEmailExists, user.Email)
		}
		return nil, erors.WrapDB(err)
	}

	// 7. Очистка чувствительных данных перед возвратом
	user.PasswordHash = ""
	return user, nil
}

// =====================================================
// UPDATE
// =====================================================

func (s *UserService) Update(ctx context.Context, id uint64, req *user_dto.UpdateUserRequest) (*models.User, error) {
	// 1. Получаем текущего пользователя
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, erors.ErrNotFound) {
			return nil, fmt.Errorf("%w: user %d", erors.ErrUserNotFound, id)
		}
		return nil, erors.WrapDB(err)
	}

	// 2. Валидация обновлений
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, fmt.Errorf("%w: %w", erors.ErrValidation, err)
	}

	// 3. Если меняется email — проверяем уникальность
	if req.Email != nil {
		newEmail := strings.ToLower(strings.TrimSpace(*req.Email))
		if newEmail != current.Email {
			if err := s.checkEmailUnique(ctx, newEmail, id); err != nil {
				return nil, err
			}
		}
	}

	// 4. Если меняется пароль — хешируем
	updates := make(map[string]interface{})
	if req.Password != nil && *req.Password != "" {
		hashed, err := hashPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", erors.ErrHashing, err)
		}
		updates["password_hash"] = hashed
	}

	// 5. Применяем остальные изменения
	applyUpdatesToMap(updates, req, current)

	// 6. Сохраняем
	if len(updates) > 0 {
		if err := s.repo.UpdatePartial(ctx, id, updates); err != nil {
			if isUniqueViolation(err, "email") {
				return nil, fmt.Errorf("%w: %s", erors.ErrEmailExists, current.Email)
			}
			return nil, erors.WrapDB(err)
		}
	}

	// 7. Возвращаем актуальную версию
	updated, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, erors.WrapDB(err)
	}
	updated.PasswordHash = ""
	return updated, nil
}

// =====================================================
// READ
// =====================================================

func (s *UserService) GetByID(ctx context.Context, id uint64) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, erors.ErrNotFound) {
			return nil, fmt.Errorf("%w: user %d", erors.ErrUserNotFound, id)
		}
		return nil, erors.WrapDB(err)
	}
	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, erors.ErrNotFound) {
			return nil, fmt.Errorf("%w: email %s", erors.ErrUserNotFound, email)
		}
		return nil, erors.WrapDB(err)
	}
	return user, nil // возвращаем с хешем для проверки пароля
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	users, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, erors.WrapDB(err)
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	return users, nil
}

// =====================================================
// DELETE
// =====================================================

func (s *UserService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, erors.ErrNotFound) {
			return fmt.Errorf("%w: user %d", erors.ErrUserNotFound, id)
		}
		return erors.WrapDB(err)
	}
	return nil
}

// TODO: Delete AUTH
// =====================================================
// AUTH
// =====================================================

// VerifyPassword проверяет пароль пользователя
func (s *UserService) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("%w: %w", erors.ErrPasswordMismatch, err)
	}
	user.PasswordHash = ""
	return user, nil
}

// =====================================================
// VALIDATION HELPERS
// =====================================================

func (s *UserService) validateCreateRequest(req *user_dto.CreateUserRequest) error {
	// Name
	if err := validateName(req.Name); err != nil {
		return fmt.Errorf("name: %w", err)
	}
	// Email
	if err := validateEmail(req.Email); err != nil {
		return fmt.Errorf("email: %w", err)
	}
	// Password
	if err := validatePassword(req.Password); err != nil {
		return fmt.Errorf("password: %w", err)
	}
	// City
	if err := validateCity(req.City); err != nil {
		return fmt.Errorf("city: %w", err)
	}
	// Role
	if req.Role != "" && req.Role != "user" && req.Role != "moderator" {
		return fmt.Errorf("role: %w", erors.ErrRoleInvalid)
	}
	// Optional fields
	if req.Faculty != "" {
		if err := validateFaculty(req.Faculty); err != nil {
			return fmt.Errorf("faculty: %w", err)
		}
	}
	if req.Course != "" {
		if err := validateCourse(req.Course); err != nil {
			return fmt.Errorf("course: %w", err)
		}
	}
	if req.About != "" && len(req.About) > 1000 {
		return fmt.Errorf("about: %w", erors.ErrAboutTooLong)
	}
	if req.AvatarImageID != "" && len(req.AvatarImageID) > 128 {
		return fmt.Errorf("avatar_image_id: %w", erors.ErrAvatarIDInvalid)
	}
	return nil
}

func (s *UserService) validateUpdateRequest(req *user_dto.UpdateUserRequest) error {
	if req.Name != nil {
		if err := validateName(*req.Name); err != nil {
			return fmt.Errorf("name: %w", err)
		}
	}
	if req.Email != nil {
		if err := validateEmail(*req.Email); err != nil {
			return fmt.Errorf("email: %w", err)
		}
	}
	if req.Password != nil && *req.Password != "" {
		if err := validatePassword(*req.Password); err != nil {
			return fmt.Errorf("password: %w", err)
		}
	}
	if req.City != nil {
		if err := validateCity(*req.City); err != nil {
			return fmt.Errorf("city: %w", err)
		}
	}
	if req.Role != nil && *req.Role != "user" && *req.Role != "moderator" {
		return fmt.Errorf("role: %w", erors.ErrRoleInvalid)
	}
	if req.Faculty != nil {
		if err := validateFaculty(*req.Faculty); err != nil {
			return fmt.Errorf("faculty: %w", err)
		}
	}
	if req.Course != nil {
		if err := validateCourse(*req.Course); err != nil {
			return fmt.Errorf("course: %w", err)
		}
	}
	if req.About != nil && len(*req.About) > 1000 {
		return fmt.Errorf("about: %w", erors.ErrAboutTooLong)
	}
	if req.AvatarImageID != nil && len(*req.AvatarImageID) > 128 {
		return fmt.Errorf("avatar_image_id: %w", erors.ErrAvatarIDInvalid)
	}
	return nil
}

// =====================================================
// FIELD VALIDATORS
// =====================================================

func validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return erors.ErrNameRequired
	}
	if len(name) < 2 {
		return erors.ErrNameTooShort
	}
	if len(name) > 255 {
		return erors.ErrNameTooLong
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '\'' {
			return erors.ErrNameInvalidChars
		}
	}
	return nil
}

func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return erors.ErrEmailRequired
	}
	if len(email) > 255 {
		return erors.ErrEmailTooLong
	}
	// Простая проверка; для продакшена используйте библиотеку
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return erors.ErrEmailInvalid
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return erors.ErrPasswordRequired
	}
	if len(password) < 8 {
		return erors.ErrPasswordTooShort
	}
	if len(password) > 128 {
		return erors.ErrPasswordTooLong
	}
	return nil
}

func validateCity(city string) error {
	city = strings.TrimSpace(city)
	if city == "" {
		return erors.ErrCityRequired
	}
	if len(city) < 2 {
		return erors.ErrCityTooShort
	}
	if len(city) > 128 {
		return erors.ErrCityTooLong
	}
	for _, r := range city {
		if !unicode.IsLetter(r) && r != ' ' && r != '-' && r != '.' {
			return erors.ErrCityTooLong // можно создать отдельную ошибку
		}
	}
	return nil
}

func validateFaculty(faculty string) error {
	faculty = strings.TrimSpace(faculty)
	if faculty == "" {
		return nil // опциональное поле
	}
	if len(faculty) > 255 {
		return erors.ErrFacultyInvalid
	}
	for _, r := range faculty {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != ' ' && r != '-' && r != '.' && r != '(' && r != ')' {
			return erors.ErrFacultyInvalid
		}
	}
	return nil
}

func validateCourse(course string) error {
	course = strings.TrimSpace(course)
	if course == "" {
		return nil // опциональное поле
	}
	if len(course) > 10 {
		return erors.ErrCourseInvalid
	}
	courseRegex := regexp.MustCompile(`^[1-6][0-9а-яА-Яa-zA-Z\-]*$`)
	if !courseRegex.MatchString(course) {
		return erors.ErrCourseInvalid
	}
	return nil
}

// =====================================================
// UTILS
// =====================================================

func (s *UserService) normalizeCreateRequest(req *user_dto.CreateUserRequest) *user_dto.CreateUserRequest {
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.City = strings.TrimSpace(req.City)
	req.About = strings.TrimSpace(req.About)
	req.Faculty = strings.TrimSpace(req.Faculty)
	req.Course = strings.TrimSpace(req.Course)
	req.AvatarImageID = strings.TrimSpace(req.AvatarImageID)
	if req.Role == "" {
		req.Role = "user"
	} else {
		req.Role = strings.ToLower(strings.TrimSpace(req.Role))
	}
	return req
}

func (s *UserService) checkEmailUnique(ctx context.Context, email string, excludeID uint64) error {
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, erors.ErrNotFound) {
			return nil // email свободен
		}
		return erors.WrapDB(err)
	}
	if existing.ID != excludeID {
		return fmt.Errorf("%w: %s", erors.ErrEmailExists, email)
	}
	return nil
}

func applyUpdatesToMap(updates map[string]interface{}, req *user_dto.UpdateUserRequest, current *models.User) {
	if req.Name != nil {
		updates["name"] = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		updates["email"] = strings.ToLower(strings.TrimSpace(*req.Email))
	}
	if req.Role != nil {
		updates["role"] = strings.ToLower(strings.TrimSpace(*req.Role))
	}
	if req.About != nil {
		updates["about"] = toNullString(strings.TrimSpace(*req.About))
	}
	if req.City != nil {
		updates["city"] = strings.TrimSpace(*req.City)
	}
	if req.Faculty != nil {
		updates["faculty"] = toNullString(strings.TrimSpace(*req.Faculty))
	}
	if req.Course != nil {
		updates["course"] = toNullString(strings.TrimSpace(*req.Course))
	}
	if req.AvatarImageID != nil {
		updates["avatar_image_id"] = toNullString(strings.TrimSpace(*req.AvatarImageID))
	}
}

func toNullString(s string) sql.NullString {
	s = strings.TrimSpace(s)
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func isUniqueViolation(err error, field string) bool {
	// Для github.com/lib/pq:
	// var pqErr *pq.Error
	// if errors.As(err, &pqErr) && pqErr.Code == "23505" { return true }

	// Универсальный фолбэк:
	return err != nil && strings.Contains(err.Error(), "unique constraint") && strings.Contains(err.Error(), field)
}
