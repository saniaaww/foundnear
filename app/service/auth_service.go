package service

import (
	"errors"
	"strings"
	"time"

	"foundnear/app/model"
	"foundnear/app/repository"
	"foundnear/helper"
	"foundnear/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo  repository.UserRepository
	roleRepo  repository.RoleRepository
	jwtSecret string
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		roleRepo:  roleRepo,
		jwtSecret: jwtSecret,
	}
}

// Register menangani pendaftaran akun baru.
func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Phone = strings.TrimSpace(req.Phone)

	if err := helper.ValidateRegister(req.Name, req.Email, req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Hash password sebelum disimpan ke database.
	passwordHash, err := helper.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to process password",
		})
	}

	// Ambil role default untuk pengguna baru.
	role, err := s.roleRepo.GetRoleByName(c.UserContext(), model.RoleUser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to get default user role",
		})
	}

	var phone *string
	if req.Phone != "" {
		phone = &req.Phone
	}

	user := &model.User{
		RoleID:       role.ID,
		RoleName:     role.Name,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Phone:        phone,
	}

	if err := s.userRepo.CreateUser(c.UserContext(), user); err != nil {
		if errors.Is(err, model.ErrDuplicate) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"status":  "error",
				"message": "Email is already registered",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to register user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Registration successful",
		"data":    toUserProfile(user),
	})
}

// Login memverifikasi kredensial dan menghasilkan JWT.
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request body",
		})
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if err := helper.ValidateLogin(req.Email, req.Password); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	user, err := s.userRepo.GetUserByEmail(c.UserContext(), req.Email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid email or password",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve user",
		})
	}

	if !helper.CheckPassword(req.Password, user.PasswordHash) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid email or password",
		})
	}

	if s.jwtSecret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "JWT configuration is missing",
		})
	}

	claims := middleware.JWTClaims{
		UserID: user.ID,
		Role:   user.RoleName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate access token",
		})
	}

	response := model.AuthResponse{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		User:        toUserProfile(user),
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Login successful",
		"data":    response,
	})
}

// Me mengambil profil pengguna berdasarkan JWT.
func (s *AuthService) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok || userID <= 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid user authentication",
		})
	}

	user, err := s.userRepo.GetUserByID(c.UserContext(), userID)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"status":  "error",
				"message": "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve user profile",
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Profile retrieved successfully",
		"data":    toUserProfile(user),
	})
}

func toUserProfile(user *model.User) model.UserProfile {
	return model.UserProfile{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.RoleName,
	}
}
