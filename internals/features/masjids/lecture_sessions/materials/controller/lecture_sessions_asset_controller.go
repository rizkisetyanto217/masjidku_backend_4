package controller

import (
	"masjidku_backend/internals/constants"
	"masjidku_backend/internals/features/masjids/lecture_sessions/materials/dto"
	"masjidku_backend/internals/features/masjids/lecture_sessions/materials/model"
	helper "masjidku_backend/internals/helpers"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// var validate = validator.New() // ✅ Buat instance validator

type LectureSessionsAssetController struct {
	DB *gorm.DB
}

func NewLectureSessionsAssetController(db *gorm.DB) *LectureSessionsAssetController {
	return &LectureSessionsAssetController{DB: db}
}

func (ctrl *LectureSessionsAssetController) CreateLectureSessionsAsset(c *fiber.Ctx) error {
	// ✅ Ambil masjid_id dari token
	masjidID, ok := c.Locals("masjid_id").(string)
	if !ok || masjidID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Masjid ID not found in token")
	}

	// ✅ Ambil field dari form
	title := c.FormValue("lecture_sessions_asset_title")
	lectureSessionID := c.FormValue("lecture_sessions_asset_lecture_session_id")

	if title == "" || lectureSessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Field wajib tidak lengkap")
	}

	// ✅ Coba ambil file upload
	var fileURL string
	var fileType int

	if file, err := c.FormFile("lecture_sessions_asset_file_url"); err == nil && file != nil {
		// ⬇️ Upload file ke Supabase
		uploadedURL, err := helper.UploadFileToSupabase("lecture_sessions_assets", file)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Gagal upload file")
		}
		fileURL = uploadedURL

		// ⬇️ Gunakan helper dari constants untuk deteksi tipe file
		fileType = constants.DetectFileTypeFromExt(file.Filename)

	} else if val := c.FormValue("lecture_sessions_asset_file_url"); val != "" {
		fileURL = val
		fileType = 1 // YouTube atau link biasa
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "Wajib upload file atau berikan URL")
	}

	// ✅ Simpan ke database
	asset := model.LectureSessionsAssetModel{
		LectureSessionsAssetTitle:            title,
		LectureSessionsAssetFileURL:          fileURL,
		LectureSessionsAssetFileType:         fileType,
		LectureSessionsAssetLectureSessionID: lectureSessionID,
		LectureSessionsAssetMasjidID:         masjidID,
	}

	if err := ctrl.DB.Create(&asset).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menyimpan asset")
	}

	return c.Status(fiber.StatusCreated).JSON(dto.ToLectureSessionsAssetDTO(asset))
}



// =============================
// 📄 Get All Assets
// =============================
func (ctrl *LectureSessionsAssetController) GetAllLectureSessionsAssets(c *fiber.Ctx) error {
	var assets []model.LectureSessionsAssetModel

	if err := ctrl.DB.Find(&assets).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve assets")
	}

	var response []dto.LectureSessionsAssetDTO
	for _, a := range assets {
		response = append(response, dto.ToLectureSessionsAssetDTO(a))
	}

	return c.JSON(response)
}



func (ctrl *LectureSessionsAssetController) GetLectureSessionsAssetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "ID tidak boleh kosong")
	}

	var asset model.LectureSessionsAssetModel
	if err := ctrl.DB.First(&asset, "lecture_sessions_asset_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Asset tidak ditemukan")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data asset")
	}

	return c.JSON(dto.ToLectureSessionsAssetDTO(asset))
}


func (ctrl *LectureSessionsAssetController) UpdateLectureSessionsAsset(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return fiber.NewError(fiber.StatusBadRequest, "ID tidak boleh kosong")
	}

	// ✅ Cari data existing
	var asset model.LectureSessionsAssetModel
	if err := ctrl.DB.First(&asset, "lecture_sessions_asset_id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "Asset tidak ditemukan")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal mengambil data asset")
	}

	// ✅ Ambil data dari form (partial)
	title := c.FormValue("lecture_sessions_asset_title")
	if title != "" {
		asset.LectureSessionsAssetTitle = title
	}

	// ✅ Cek apakah ada file baru diupload
	file, errFile := c.FormFile("lecture_sessions_asset_file_url")
	if errFile == nil && file != nil {
		// Hapus file lama jika dari Supabase
		if asset.LectureSessionsAssetFileURL != "" {
			if parsed, err := url.Parse(asset.LectureSessionsAssetFileURL); err == nil {
				prefix := "/storage/v1/object/public/"
				cleaned := strings.TrimPrefix(parsed.Path, prefix)

				if unescaped, err := url.QueryUnescape(cleaned); err == nil {
					parts := strings.SplitN(unescaped, "/", 2)
					if len(parts) == 2 {
						_ = helper.DeleteFromSupabase(parts[0], parts[1])
					}
				}
			}
		}

		// Upload file baru
		uploadedURL, err := helper.UploadFileToSupabase("lecture_sessions_assets", file)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Gagal upload file")
		}

		asset.LectureSessionsAssetFileURL = uploadedURL
		asset.LectureSessionsAssetFileType = constants.DetectFileTypeFromExt(file.Filename)

	} else if val := c.FormValue("lecture_sessions_asset_file_url"); val != "" {
		asset.LectureSessionsAssetFileURL = val
		asset.LectureSessionsAssetFileType = 1 // YouTube atau link biasa
	}

	// ✅ Simpan update
	if err := ctrl.DB.Save(&asset).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal memperbarui asset")
	}

	return c.JSON(dto.ToLectureSessionsAssetDTO(asset))
}

func (ctrl *LectureSessionsAssetController) DeleteLectureSessionsAsset(c *fiber.Ctx) error {
	id := c.Params("id")

	// 🔍 Cek apakah asset ada
	var asset model.LectureSessionsAssetModel
	if err := ctrl.DB.First(&asset, "lecture_sessions_asset_id = ?", id).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Asset tidak ditemukan")
	}

	// 🗑️ Hapus file dari Supabase jika file URL adalah dari Supabase
	if asset.LectureSessionsAssetFileURL != "" {
		parsed, err := url.Parse(asset.LectureSessionsAssetFileURL)
		if err == nil {
			rawPath := parsed.Path // /storage/v1/object/public/file/lecture_sessions_assets%2Fxxx.pdf
			prefix := "/storage/v1/object/public/"
			cleaned := strings.TrimPrefix(rawPath, prefix)

			unescaped, err := url.QueryUnescape(cleaned)
			if err == nil {
				parts := strings.SplitN(unescaped, "/", 2)
				if len(parts) == 2 {
					bucket := parts[0]
					objectPath := parts[1]
					_ = helper.DeleteFromSupabase(bucket, objectPath)
				}
			}
		}
	}

	// ❌ Hapus dari DB
	if err := ctrl.DB.Delete(&asset).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Gagal menghapus asset")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Asset berhasil dihapus",
	})
}
