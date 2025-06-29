package batch_test

import (
	"testing"

	"github.com/ryabkov82/shortener/internal/app/service/mocks"

	"github.com/ryabkov82/shortener/test/testhandlers"

	"github.com/golang/mock/gomock"
)

func TestGetHandler(t *testing.T) {

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockRepository(ctrl)

	m.EXPECT().GetExistingURLs(gomock.Any(), gomock.Any()).Return(nil, nil)
	m.EXPECT().SaveNewURLs(gomock.Any(), gomock.Any()).Return(nil)

	testhandlers.TestBatch(t, m)
}
