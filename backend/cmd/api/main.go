package main

import (
	"backend/backend/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"backend/backend/helper"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

//Connection mongoDB with helper class
var collection1, collection2 = helper.ConnectDB()

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

var SECRET_KEY = []byte("gosecretkey")

func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Error in JWT token generation")
		return "", err
	}
	return tokenString, nil
}

func userSignup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user models.User
	json.NewDecoder(request.Body).Decode(&user)
	user.Password = getHash([]byte(user.Password))
	//collection := client.Database("GODB").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	result, _ := collection1.InsertOne(ctx, user)
	json.NewEncoder(response).Encode(result)
}

func userLogin(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	var user models.User
	var dbUser models.User
	json.NewDecoder(request.Body).Decode(&user)
	//collection:= client.Database("GODB").Collection("user")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := collection1.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	userPass := []byte(user.Password)
	dbPass := []byte(dbUser.Password)

	passErr := bcrypt.CompareHashAndPassword(dbPass, userPass)

	if passErr != nil {
		log.Println(passErr)
		response.Write([]byte(`{"response":"Wrong Password!"}`))
		return
	}
	jwtToken, err := GenerateJWT()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"` + err.Error() + `"}`))
		return
	}
	response.Write([]byte(`{"token":"` + jwtToken + `"}`))

}

func getBlogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// we created Blog array
	var Blogs []models.Blog

	// bson.M{},  we passed empty filter. So we want to get all data.
	cur, err := collection2.Find(context.TODO(), bson.M{})

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close the cursor once finished
	/*A defer statement defers the execution of a function until the surrounding function returns.
	simply, run cur.Close() process but after cur.Next() finished.*/
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		// create a value into which the single document can be decoded
		var Blog models.Blog
		// & character returns the memory address of the following variable.
		err := cur.Decode(&Blog) // decode similar to deserialize process.
		if err != nil {
			log.Fatal(err)
		}

		// add item our array
		Blogs = append(Blogs, Blog)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	json.NewEncoder(w).Encode(Blogs) // encode similar to serialize process.
}

func getBlog(w http.ResponseWriter, r *http.Request) {
	// set header.
	w.Header().Set("Content-Type", "application/json")

	var Blog models.Blog
	// we get params with mux.
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// We create filter. If it is unnecessary to sort data for you, you can use bson.M{}
	filter := bson.M{"_id": id}
	err := collection2.FindOne(context.TODO(), filter).Decode(&Blog)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(Blog)
}

func createBlog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var Blog models.Blog

	// we decode our body request params
	_ = json.NewDecoder(r.Body).Decode(&Blog)

	// insert our Blog model.
	result, err := collection2.InsertOne(context.TODO(), Blog)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func updateBlog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var params = mux.Vars(r)

	//Get id from parameters
	id, _ := primitive.ObjectIDFromHex(params["id"])

	var Blog models.Blog

	// Create filter
	filter := bson.M{"_id": id}

	// Read update model from body request
	_ = json.NewDecoder(r.Body).Decode(&Blog)

	// prepare update model.
	update := bson.D{
		{"$set", bson.D{
			{"tags", Blog.Tags},
			{"title", Blog.Title},
			//{"user", bson.D{
			//	{"name", Blog.User.Name},
			//{"email", Blog.User.Email},
		}},
	}

	err := collection2.FindOneAndUpdate(context.TODO(), filter, update).Decode(&Blog)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	Blog.ID = id

	json.NewEncoder(w).Encode(Blog)
}

func deleteBlog(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")

	// get params
	var params = mux.Vars(r)

	// string to primitve.ObjectID
	id, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"_id": id}

	deleteResult, err := collection2.DeleteOne(context.TODO(), filter)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)

}

// search user with email id
func getBlogUser(w http.ResponseWriter, r *http.Request) {
	// Set header
	w.Header().Set("Content-Type", "application/json")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// get params
	var params = mux.Vars(r)
	fmt.Println(params["email"])
	// string to primitve.ObjectID
	//ID, err := primitive.ObjectIDFromHex(params["id"])

	// prepare filter.
	filter := bson.M{"title": params["email"]}

	Result, err := collection2.Find(context.TODO(), filter)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	var arr []bson.M
	if err = Result.All(ctx, &arr); err != nil {
		log.Fatal(err)
	}
	fmt.Println(arr)

	json.NewEncoder(w).Encode(arr)
}

// var client *mongo.Client

func main() {
	//Init Router
	r := mux.NewRouter()

	r.HandleFunc("/api/Blogs", getBlogs).Methods("GET")
	r.HandleFunc("/api/Blogs/{id}", getBlog).Methods("GET")
	r.HandleFunc("/api/Blogs", createBlog).Methods("POST")
	r.HandleFunc("/api/Blogs/{id}", updateBlog).Methods("PUT")
	r.HandleFunc("/api/Blogs/{id}", deleteBlog).Methods("DELETE")
	r.HandleFunc("/api/user/signup", userSignup).Methods("POST")
	r.HandleFunc("/api/user/login", userLogin).Methods("POST")
	r.HandleFunc("/api/users/Blogs/{email}", getBlogUser).Methods("GET")

	//config := helper.GetConfiguration()
	log.Fatal(http.ListenAndServe(":8000", r))

}
