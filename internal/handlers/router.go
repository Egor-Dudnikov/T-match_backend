// Copyright (c) 2026 Egor Dudnikov
// SPDX-License-Identifier: MIT

package handlers

import (
	"net/http"

	"T-match_backend/internal/constants"

	"github.com/julienschmidt/httprouter"
)

func NewRouter(app *ServiceHandler) *httprouter.Router {
	router := httprouter.New()

	router.GET("/", ErrorMiddleware(
		app.CorsMiddleware(app.Index)))

	// Auth routes
	router.POST("/auth/students", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthStudentHandler, constants.RateLimitAuthStudent, "/auth/students"))))

	router.POST("/auth/students/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyUserHandler, constants.RateLimitAuthStudentVerify, "/auth/students/verify"))))

	router.POST("/auth/newverify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.NewVerifyCode, constants.RateLimitNewVerifyCode, "/auth/newverify"))))

	router.POST("/auth/students/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginUserHandler, constants.RateLimitAuthStudentLogin, "/auth/students/login"))))

	router.POST("/auth/company", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.AuthCompanyHandler, constants.RateLimitAuthCompany, "/auth/company"))))

	router.POST("/auth/company/verify", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.VerifyCompanyHandler, constants.RateLimitAuthCompanyVerify, "/auth/company/verify"))))

	router.POST("/auth/company/login", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.LoginCompanyHandler, constants.RateLimitAuthCompanyLogin, "/auth/company/login"))))

	// Profile routes
	router.PUT("/my/profile/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.UpdateProfileHandler, constants.RateLimitUpdateProfile, "/my/profile/put"))))))

	router.GET("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.GetMyProfileHandler, constants.RateLimitGetProfile, "my/profile"))))))

	router.POST("/my/profile/skills/add", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.AddInternSkillsHandler, constants.RateLimitAddSkills, "/my/profile/skills/add"))))))

	router.DELETE("/my/profile/skills/delete", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.DeleteInternSkillsHandler, constants.RateLimitDeleteSkills, "/my/profile/skills/delete"))))))

	router.PUT("/my/company/profile/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateCompanyProfileHandler, constants.RateLimitUpdateCompany, "/my/company/profile"))))))

	router.GET("/my/company/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.GetMyCompanyProfileHandler, constants.RateLimitGetCompany, "/my/company/profile"))))))

	router.PUT("/my/avatar/put", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.RateLimitMiddleware(app.SetMyAvatarHandler, constants.RateLimitSetAvatar, "/my/avatar/put")))))

	// Internship routes
	router.POST("/internships", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.NewInternshipHandler, constants.RateLimitCreateInternship, "/internships"))))))

	router.GET("/internships/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.GetInternshipByIdHandler)))

	router.PUT("/internships/update/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateInternshipHandler, constants.RateLimitUpdateInternship, "/internships/update/"))))))

	router.DELETE("/internships/delete/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.ArchivedInternshipHandler, constants.RateLimitArchiveInternship, "/internships/delete/"))))))

	router.GET("/skills", ErrorMiddleware(
		app.CorsMiddleware(app.GetAllSkills)))

	router.POST("/internship/:id/skill/add", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.AddInternshipSkillsHandler, constants.RateLimitAddInternshipSkill, "/internship/skill/add"))))))

	router.DELETE("/internship/:id/skill/delete", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.DeleteInternshipSkillsHandler, constants.RateLimitDeleteInternshipSkill, "/internship/skill/delete"))))))

	router.POST("/internships/:id/respond", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.RespondInternshipHandler, constants.RateLimitRespondToInternship, "/internships/:id/respond"))))))

	router.GET("/my/responses", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.GetMyResponsesHandler, constants.RateLimitGetMyResponses, "/my/responses"))))))

	router.GET("/internships/:id/responses", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.GetInternshipResponses, constants.RateLimitGetInternshipResponses, "/internships/:id/responses"))))))

	router.PUT("/responses/:id/status", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.SetResponseStatus, constants.RateLimitSetResponseStatus, "/responses/:id/status"))))))

	// Search routes
	router.GET("/internships", ErrorMiddleware(
		app.CorsMiddleware(app.SearchInternshipHandler)))

	router.GET("/companies", ErrorMiddleware(
		app.CorsMiddleware(app.SearchCompanyHandler)))

	router.GET("/students", ErrorMiddleware(
		app.CorsMiddleware(app.SearchInternHandler)))

	// CORS preflight
	router.OPTIONS("/*path", ErrorMiddleware(app.CorsMiddleware(handleOptions)))

	return router
}

// handleOptions handles CORS preflight requests
func handleOptions(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
