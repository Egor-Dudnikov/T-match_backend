// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *ServiceHandler) *httprouter.Router {
	router := httprouter.New()

	router.GET("/", ErrorMiddleware(
		app.CorsMiddleware(app.Index)))

	router.POST("/auth/students", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthStudentHandler, 20, "/auth/students"))))

	router.POST("/auth/students/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyUserHandler, 60, "/auth/students/verify"))))

	router.POST("/auth/newverify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.NewVerifyCode, 7, "/auth/newverify"))))

	router.POST("/auth/students/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginUserHandler, 30, "/auth/students/login"))))

	router.POST("/auth/company", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthCompanyHandler, 20, "/auth/company"))))

	router.POST("/auth/company/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyCompanyHandler, 60, "/auth/company/verify"))))

	router.POST("/auth/company/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginCompanyHandler, 30, "/auth/company/login"))))

	router.PUT("/my/profile/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.UpdateProfileHandler, 100, "/my/profile/put"))))))

	router.GET("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.GetMyProfileHandler, 120, "my/profile"))))))

	router.POST("/my/profile/skills/add", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.AddInternSkillsHandler, 5, "/my/profile/skills/add"))))))

	router.DELETE("/my/profile/skills/delete", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.DeleteInternSkillsHandler, 5, "/my/profile/skills/delete"))))))

	router.PUT("/my/company/profile/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateCompanyProfileHandler, 100, "/my/company/profile"))))))

	router.GET("/my/company/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.GetMyCompanyProfileHandler, 120, "/my/company/profile"))))))

	router.PUT("/my/avatar/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.RateLimitMiddleware(app.SetMyAvatarHandler, 100, "/my/avatar/put")))))

	router.POST("/internships", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.NewInternshipHandler, 12, "/internships"))))))

	router.GET("/internships/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.GetInternshipByIdHandler)))

	router.PUT("/internships/update/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateInternshipHandler, 12, "/internships/update/"))))))

	router.DELETE("/internships/delete/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.ArchivedInternshipHandler, 5, "/internships/delete/"))))))

	router.GET("/skills", ErrorMiddleware(
		app.CorsMiddleware(app.GetAllSkills)))

	router.POST("/internship/:id/skill/add", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.AddInternshipSkillsHandler, 5, "/internship/skill/add"))))))

	router.DELETE("/internship/:id/skill/delete", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.DeleteInternshipSkillsHandler, 5, "/internship/skill/delete"))))))

	router.POST("/internships/:id/respond", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.RespondInternshipHandler, 10, "/internships/:id/respond"))))))

	router.GET("/my/responses", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.GetMyResponsesHandler, 20, "/my/responses"))))))

	router.GET("/internships/:id/responses", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.GetInternshipResponses, 20, "/internships/:id/responses"))))))

	router.PUT("/responses/:id/status", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.SetResponseStatus, 20, "/responses/:id/status"))))))

	router.GET("/internships", ErrorMiddleware(
		app.CorsMiddleware(app.SearchInternshipHandler)))

	router.GET("/companies", ErrorMiddleware(
		app.CorsMiddleware(app.SearchCompanyHandler)))

	router.GET("/students", ErrorMiddleware(
		app.CorsMiddleware(app.SearchInternHandler)))

	router.OPTIONS("/*path", ErrorMiddleware(app.CorsMiddleware(handleOptions)))

	return router
}

// заглушка для OPTIONS

func handleOptions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
