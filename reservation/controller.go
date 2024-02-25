package reservation

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/shkuran/go-library/book"
	"github.com/shkuran/go-library/utils"
)

type ReservationController interface {
	GetReservations(context *gin.Context)
	AddReservation(context *gin.Context)
	CompleteReservation(context *gin.Context)
}

type ReservationControllerImpl struct {
	repo Repository
}

func NewReservationController(repo Repository) *ReservationControllerImpl {
	return &ReservationControllerImpl{repo: repo}
}

func (rsv ReservationControllerImpl) GetReservations(context *gin.Context) {
	reservations, err := rsv.repo.GetReservations()
	if err != nil {
		utils.HandleInternalServerError(context, "Could not fetch reservations!", err)
		return
	}

	context.JSON(http.StatusOK, reservations)
}

func (rsv ReservationControllerImpl) AddReservation(context *gin.Context) {
	var reservation Reservation
	err := context.ShouldBindJSON(&reservation)
	if err != nil {
		utils.HandleBadRequest(context, "Could not parse request data!", err)
		return
	}

	b, err := book.GetBookById(reservation.BookId)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not fetch book!", err)
		return
	}

	numberOfBookCopies := b.AvailableCopies
	if numberOfBookCopies < 1 {
		utils.HandleBadRequest(context, "The book is not available!", nil)
		return
	}

	userId := context.GetInt64("userId")
	reservation.UserId = userId

	err = rsv.repo.SaveReservation(reservation)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not add reservation!", err)
		return
	}

	err = book.UpdateNumberOfBooks(b.ID, b.AvailableCopies-1)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not update the number of book copies!", err)
		return
	}

	utils.HandleStatusCreated(context, "Reservation added!")
}

func (rsv *ReservationControllerImpl) CompleteReservation(context *gin.Context) {
	reservationId, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		utils.HandleBadRequest(context, "Could not parse reservationId!", err)
		return
	}

	reservation, err := rsv.repo.GetReservationById(reservationId)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not fetch reservation!", err)
		return
	}

	userId := context.GetInt64("userId")
	if reservation.UserId != userId {
		utils.HandleStatusUnauthorized(context, "Not access to copmlete reservation!", nil)
		return
	}

	if reservation.ReturnDate != nil {
		utils.HandleBadRequest(context, "The reservation is copleted already!!", nil)
		return
	}

	err = rsv.repo.UpdateReturnDate(reservationId)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not copmlete reservation!", err)
		return
	}

	b, err := book.GetBookById(reservation.BookId)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not fetch book!", err)
		return
	}

	err = book.UpdateNumberOfBooks(b.ID, b.AvailableCopies+1)
	if err != nil {
		utils.HandleInternalServerError(context, "Could not update the number of book copies!", err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Reservation copmleted!"})
}
