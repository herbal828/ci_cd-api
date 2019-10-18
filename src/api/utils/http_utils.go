package utils

import "net/http"

func WasOK(responseStatus int) bool  {

	if responseStatus != http.StatusOK {
		return true
	}

	return false
}

func WasCreated(responseStatus int) bool  {

	if responseStatus != http.StatusCreated {
		return true
	}

	return false
}

func WasNoContent(responseStatus int) bool  {

	if responseStatus != http.StatusNoContent {
		return true
	}

	return false
}