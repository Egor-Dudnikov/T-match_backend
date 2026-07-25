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
		app.CorsMiddleware(app.CheckHealth)))

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

	router.POST("/auth/logout", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(app.LogoutHandler))))

	router.POST("/auth/forgot-password", ErrorMiddleware(
		app.CorsMiddleware(app.ForgotPasswordHandler)))

	router.POST("/auth/forgot-password/verify", ErrorMiddleware(
		app.CorsMiddleware(app.VerifyForgotPasswordHandler)))

	router.PUT("/auth/change-password", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(app.ChangePasswordHandler))))

	// Profile routes
	router.PUT("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.UpdateProfileHandler, constants.RateLimitUpdateProfile, "/my/profile"))))))

	router.GET("/my/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.GetMyProfileHandler, constants.RateLimitGetProfile, "/my/profile"))))))

	router.POST("/my/profile/skills", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.AddInternSkillsHandler, constants.RateLimitAddSkills, "/my/profile/skills"))))))

	router.DELETE("/my/profile/skills", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.DeleteInternSkillsHandler, constants.RateLimitDeleteSkills, "/my/profile/skills"))))))

	router.GET("/company/profile/:id", ErrorMiddleware(
		app.CorsMiddleware(app.GetCompanyProfileHandler)))

	router.GET("/profile/:id", ErrorMiddleware(
		app.CorsMiddleware(app.GetProfileHandler)))

	router.PUT("/my/company/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateCompanyProfileHandler, constants.RateLimitUpdateCompany, "/my/company/profile"))))))

	router.GET("/my/company/profile", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.GetMyCompanyProfileHandler, constants.RateLimitGetCompany, "/my/company/profile"))))))

	router.PUT("/my/avatar", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.RateLimitMiddleware(app.SetMyAvatarHandler, constants.RateLimitSetAvatar, "/my/avatar")))))

	// Internship routes

	router.GET("/internships", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.SearchInternshipHandler, constants.RateLimitSearchInternship, "/internships"))))

	router.POST("/internships", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.NewInternshipHandler, constants.RateLimitCreateInternship, "/internships"))))))

	router.GET("/internships/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.GetInternshipByIDHandler)))

	router.PUT("/internships/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.UpdateInternshipHandler, constants.RateLimitUpdateInternship, "/internships/put"))))))

	router.DELETE("/internships/:id", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.ArchivedInternshipHandler, constants.RateLimitArchiveInternship, "/internships/delete"))))))

	router.GET("/skills", ErrorMiddleware(
		app.CorsMiddleware(app.GetAllSkills)))

	router.POST("/internships/:id/skill", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.AddInternshipSkillsHandler, constants.RateLimitAddInternshipSkill, "/internship/skill"))))))

	router.DELETE("/internships/:id/skill", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.CompanyMiddleware(
					app.RateLimitMiddleware(app.DeleteInternshipSkillsHandler, constants.RateLimitDeleteInternshipSkill, "/internship/skill"))))))

	router.POST("/internships/:id/respond", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.RespondInternshipHandler, constants.RateLimitRespondToInternship, "/internships/:id/respond"))))))

	router.DELETE("/internships/:id/respond", ErrorMiddleware(
		app.CorsMiddleware(
			app.AuthMiddleware(
				app.InternMiddleware(
					app.RateLimitMiddleware(app.DeleteRespondInternshipHandler, constants.RateLimitRespondToInternship, "/internships/:id/respond"))))))

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

	router.GET("/companies", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.SearchCompanyHandler, constants.RateLimitSearchCompany, "/companies"))))

	router.GET("/students", ErrorMiddleware(
		app.CorsMiddleware(
			app.RateLimitMiddleware(app.SearchInternHandler, constants.RateLimitSearchStudent, "/students"))))

	// CORS preflight
	router.OPTIONS("/*path", ErrorMiddleware(app.CorsMiddleware(handleOptions)))

	return router
}

// handleOptions handles CORS preflight requests
func handleOptions(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
