package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ariopri/Let-It-Be/tree/main/backend/entities"
	"github.com/ariopri/Let-It-Be/tree/main/backend/handler/middleware"
	"github.com/ariopri/Let-It-Be/tree/main/backend/utils/token"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type userHandler struct {
	userRepo entities.UserRepository
}

// routes
func NewUserHandler(r *gin.Engine, userRepo entities.UserRepository) {
	handler := &userHandler{
		userRepo: userRepo,
	}

	// middleware
	m := middleware.InitMiddleware()
	auth := r.Group("/api").Use(m.JWTMiddleware())
	{
		auth.GET("/users", handler.fetch)
		auth.GET("/users/:id", handler.fetchById)
		auth.POST("/users", handler.create)
		auth.PUT("/users/:id", handler.update)
		auth.DELETE("/users/:id", handler.delete)
	}

	// should be public routes
	r.POST("/login", handler.login)
	r.POST("/register", handler.register)
}

func errMessage(v validator.FieldError) string {
	m := fmt.Sprintf("error on field %s, condition: %s", v.Field(), v.ActualTag())

	return m
}

// login
func (u *userHandler) login(c *gin.Context) {
	ctx := c.Request.Context()
	var login entities.Login

	if err := c.ShouldBind(&login); err != nil {
		for _, v := range err.(validator.ValidationErrors) {
			eM := errMessage(v)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": eM,
			})

			return
		}
	}

	userLogin, err := u.userRepo.Login(ctx, &login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})

		return
	}

	// JWT
	token, _ := token.CreateToken(userLogin.Email, userLogin.Role)

	c.JSON(http.StatusOK, gin.H{
		"message": "user logged in",
		"token":   token,
		"data":    userLogin,
	})
}

// register
func (u *userHandler) register(c *gin.Context) {
	ctx := c.Request.Context()
	user := entities.User{}

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": entities.BadRequest,
		})
		return
	}

	userData, err := u.userRepo.Register(ctx, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})
		return
	}

	// JWT
	token, _ := token.CreateToken(userData.Email, userData.Role)

	c.JSON(http.StatusOK, gin.H{
		"message": "user registered",
		"data":    userData,
		"token":   token,
	})
}

// fetch users
func (u *userHandler) fetch(c *gin.Context) {
	ctx := c.Request.Context()
	users, err := u.userRepo.Fetch(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})

		return
	}

	// role check
	auth := c.Request.Header.Get("Authorization")

	token, _ := token.ValidateToken(auth)

	if token.Role != "admin" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": entities.Unauthorized,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "users fetched",
		"users":   users,
	})
}

// fetch user by id
func (u *userHandler) fetchById(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idConv, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": entities.BadRequest,
		})
		return
	}

	user, err := u.userRepo.FetchById(ctx, int64(idConv))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user fetched",
		"user":    user,
	})
}

// create user
func (u *userHandler) create(c *gin.Context) {
	ctx := c.Request.Context()
	user := entities.User{}

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": entities.BadRequest,
		})
		return
	}

	userData, err := u.userRepo.Create(ctx, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user created",
		"data":    userData,
	})
}

// update user
func (u *userHandler) update(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	idConv, _ := strconv.Atoi(id)
	user := entities.User{}

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": entities.BadRequest,
		})
		return
	}

	userData, err := u.userRepo.Update(ctx, int64(idConv), &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.InternalServer,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user updated",
		"user":    userData,
	})
}

// delete user
func (u *userHandler) delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	idConv, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": entities.BadRequest,
		})
		return
	}

	if err := u.userRepo.Delete(ctx, int64(idConv)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": entities.ItemNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user deleted",
	})
}
