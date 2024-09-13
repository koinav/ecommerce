package database

import "errors"

var (
	ErrCantFindProduct        = errors.New("cannot find product")
	ErrCantDecodeProducts     = errors.New("cannot find product")
	ErrUserIdIsNotValid       = errors.New("this user is not valid")
	ErrCantUpdateUser         = errors.New("cannot add this product to the cart")
	ErrCantRemoveItemFromCart = errors.New("cannot remove this item from the cart")
	ErrCantGetItem            = errors.New("unable to get the item from the cart")
	ErrCantBuyCartItem        = errors.New("cannot update the purchase")
)

func AddProductToCart() {

}

func RemoveCartItem() {

}

func BuyItemFromCart() {

}

func InstantBuy() {

}
