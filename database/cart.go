package database

import (
	"context"
	"errors"
	"github.com/koinav/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

var (
	ErrCantFindProduct        = errors.New("cannot find product")
	ErrCantDecodeProducts     = errors.New("cannot find product")
	ErrUserIdIsNotValid       = errors.New("this user is not valid")
	ErrCantUpdateUser         = errors.New("cannot add this product to the cart")
	ErrCantRemoveItemFromCart = errors.New("cannot remove this item from the cart")
	ErrCantGetItem            = errors.New("unable to get the item from the cart")
	ErrCantBuyCartItem        = errors.New("cannot update the purchase")
)

func AddProductToCart(ctx context.Context,
	productCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID, userID string) error {
	searchFromDB, err := productCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.ProductInCart
	err = searchFromDB.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{
		Key: "$push",
		Value: bson.D{primitive.E{
			Key:   "user_cart",
			Value: bson.D{{Key: "$each", Value: productCart}},
		}},
	}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}

	return nil
}

func RemoveCartItem(ctx context.Context,
	productCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"user_cart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return ErrCantRemoveItemFromCart
	}

	return nil
}

func BuyItemFromCart(ctx context.Context,
	userCollection *mongo.Collection, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var getCartItems models.User
	var orderCart models.Order

	orderCart.OrderID = primitive.NewObjectID()
	orderCart.OrderedAt = time.Now()
	orderCart.OrderCart = make([]models.ProductInCart, 0)
	orderCart.PaymentMethod.COD = true

	unwind := bson.D{{Key: "&unwind", Value: bson.D{primitive.E{Key: "path", Value: "$user_cart"}}}}
	grouping := bson.D{{
		Key: "$group",
		Value: bson.D{primitive.E{Key: "_id", Value: "$_id"},
			{Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$user_cart.price"}}}},
	}}

	currentRes, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	if err != nil {
		log.Println(err)
		return err
	}

	var getUserCart []bson.M
	if err = currentRes.All(ctx, &getUserCart); err != nil {
		log.Println(err)
		return err
	}

	var totalPrice int
	for _, item := range getUserCart {
		price := item["total"]
		totalPrice = price.(int)
	}
	orderCart.Price = totalPrice

	filter := bson.D{primitive.E{Key: "_id", Value: "id"}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderCart}}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return err
	}

	err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
		return err
	}

	filter1 := bson.D{primitive.E{Key: "_id", Value: id}}
	update1 := bson.M{"&push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}
	_, err = userCollection.UpdateOne(ctx, filter1, update1)
	if err != nil {
		log.Println(err)
		return err
	}

	userCartEmpty := make([]models.ProductInCart, 0)
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "user_cart", Value: userCartEmpty}}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func InstantBuy(ctx context.Context,
	productCollection, userCollection *mongo.Collection,
	productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	var productDetails models.ProductInCart
	var orderDetails models.Order

	orderDetails.OrderID = primitive.NewObjectID()
	orderDetails.OrderedAt = time.Now()
	orderDetails.OrderCart = make([]models.ProductInCart, 0)

	orderDetails.PaymentMethod.COD = true
	err = productCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&productDetails)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	orderDetails.Price = productDetails.Price

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orderDetails}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	filter1 := bson.D{primitive.E{Key: "_id", Value: id}}
	update1 := bson.M{"&push": bson.M{"orders.$[].order_list": productDetails}}
	_, err = userCollection.UpdateOne(ctx, filter1, update1)
	if err != nil {
		log.Println(err)
	}

	return nil
}
