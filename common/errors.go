package common

import "errors"

var (
	ErrSlugAlreadyExists = errors.New("slug đã tồn tại")

	ErrCategoryNotFound = errors.New("không tìm thấy danh mục sản phẩm")

	ErrHasCategoryNotFound = errors.New("có danh mục sản phẩm không tìm thấy")

	ErrHasTagNotFound = errors.New("có tag mục sản phẩm không tìm thấy")

	ErrProductNotFound = errors.New("không tìm thấy sản phẩm")

	ErrColorAlreadyExists = errors.New("màu sắc đã tồn tại")

	ErrSizeAlreadyExists = errors.New("size đã tồn tại")

	ErrSKUAlreadyExists = errors.New("SKU đã tồn tại")

	ErrColorNotFound = errors.New("không tìm thấy màu sắc")

	ErrSizeNotFound = errors.New("không tìm thấy kích cỡ")

	ErrUnSupportedFileType = errors.New("định dạng file không được hỗ trợ")

	ErrTagAlreadyExists = errors.New("tag sản phẩm đã tồn tại")

	ErrTagNotFound = errors.New("không tìm thấy tag sản phẩm")

	ErrHasImageNotFound = errors.New("có ảnh sản phẩm không tìm thấy")

	ErrImageNotFound = errors.New("không tìm thấy hình ảnh sản phẩm")

	ErrHasVariantNotFound = errors.New("có biến thể sản phẩm không tìm thấy")

	ErrVariantNotFound = errors.New("không tìm thấy biến thể sản phẩm")

	ErrInventoryNotFound = errors.New("không tìm thấy tồn kho biến thể")

	ErrHasProductNotFound = errors.New("có sản phẩm không tìm thấy")

	ErrHasColorNotFound = errors.New("có màu sắc không tìm thấy")

	ErrHasSizeNotFound = errors.New("có kích cỡ không tìm thấy")

	ErrHasSKUAlreadyExists = errors.New("có SKU đã tồn tại")

	ErrHasUserNotFound = errors.New("có người dùng không tìm thấy")

	ErrUserNotFound = errors.New("không tìm thấy người dùng")
)