package route

import (
	"masjidku_backend/internals/features/masjids/lectures/main/controller"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AllLectureRoutes(api fiber.Router, db *gorm.DB) {
	lectureCtrl := controller.NewLectureController(db)

	lecture := api.Group("/lectures")
	lecture.Get("/", lectureCtrl.GetAllLectures)
	lecture.Get("/:id/lecture-sessions", lectureCtrl.GetLectureSessionsByLectureID)

	lecture.Get("/:id", lectureCtrl.GetLectureByID)
	lecture.Get("/slug/:slug", lectureCtrl.GetLectureByMasjidSlug)

	ctrl := controller.NewUserLectureController(db)

	userLecture := api.Group("/user-lectures")
	userLecture.Post("/", ctrl.CreateUserLecture)
	userLecture.Post("/by-lecture", ctrl.GetUsersByLecture) // ✅ opsional tambahan jika ingin ambil semua kajian yang diikuti user
	// 📌 Get progress for logged in user (user_id from JWT token)
	userLecture.Get("/with-progress", ctrl.GetUserLectureStats)

	// 🧑‍🏫 Route untuk user_lecture_sessions
	userLectureSessions := api.Group("/user-lecture-sessions-in-lecture")
	userLectureSessions.Get("/with-progress", ctrl.GetUserLecturesSessionsInLectureWithProgress)

}
