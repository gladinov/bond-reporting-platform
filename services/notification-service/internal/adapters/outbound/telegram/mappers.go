package telegram

import "github.com/gladinov/notification-service/internal/application/usecases"

func mapUsecaseImageToTgImage(image *usecases.ImageData) *ImageData {
	var res ImageData
	res.Name = image.Name
	res.Data = image.Data
	res.Caption = image.Caption
	return &res
}

func mapUsecaseImagesToTgImages(images []*usecases.ImageData) []*ImageData {
	res := make([]*ImageData, 0, len(images))
	for i := range images {
		res = append(res, mapUsecaseImageToTgImage(images[i]))
	}
	return res
}
