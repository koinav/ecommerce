# Ecommerce Go

Готовое (почти...) API-решение для интернет-магазина. Проект включает в себя масштабируемую архитектуру для управления товарами, клиентами и заказами в рамках онлайн-платформы для продаж.

## 📚 Используемые технологии

Проект написан с использованием следующих технологий:

- **Gin Web Framework**: Легковесный и быстрый HTTP web framework для Go (Golang).
- **MongoDB**: NoSQL база данных для хранения данных о пользователях и товарах.
- **Docker**: Платформа для разработки, доставки и запуска приложений в контейнерах.

## 🚀 Запуск приложения

### Шаги для запуска:

1. **Клонирование репозитория**:
```bash
   git clone https://github.com/koinav/ecommerce.git
   cd ecommerce
```

2. **Инициализация Docker-контейнеров**:
```bash
docker-compose up -d
```

3. **Запуск приложения на go**:
```bash
go run main.go
```

_Теперь API готово к использованию._

Вы можете получить к нему доступ по адресу: `http://localhost:8000`.

## 📄 Функциональность

Вот основные функции, которые выполняет проложение:

- **Управление профилем клиента**: Регистрация пользователя, изменение информации об адресе доставки и способах оплаты.
- **Оформление заказов**: Управление корзиной покупок и оформление заказов клиента.
- **Управление товарами**: Добавление и удаление информации о товарах (для админисраторов).

Далее детально.

### API-вызовы, доступные без регистрации пользователя

- **SignUp (POST)** _[регистрация]_

http://localhost:8000/users/signup

```json
{
  "first_name": "Zubenko",
  "last_name": "Mikhail",
  "email": "zubenko@yandex.ru",
  "password": "NewPass",
  "phone": "+79991234353"
}
```

Если все успешно, ответ: "Successfully signed up!"

- **LogIn (POST)** _[аутентификация]_

  http://localhost:8000/users/login

```json
{
  "email": "zubenko@yandex.ru",
  "password": "NewPass"
}
```

Ответ будет следующего вида:

```json
{
  "_id": "***********************",
  "first_name": "Zubenko",
  "last_name": "Mikhail",
  "password": "$2a$14$9OfB7vopeDdUZdGj7DP/l..//i4qw1IPdxDSkIHFEAhYsnIQ9d0XK",
  "email": "zubenko@gmail.com",
  "phone": "+79991234353",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6ImlpdmFub3ZAeWFuZGV4LnJ1IiwiRmlyc3ROYW1lIjoiSXZhbiIsIkxhc3ROYW1lIjoiSXZhbm92IiwiVWlkIjoiNjZlNmQ2MDBlZDFlMTBkZWRjM2RiMGIxIiwiZXhwIjoxNzI2NDkwNDk2fQ.yp3pDNKTkOMn6gte6cU7UJRzL-i6P6qcVxdOQtPZz_c",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6IiIsIkZpcnN0TmFtZSI6IiIsIkxhc3ROYW1lIjoiIiwiVWlkIjoiIiwiZXhwIjoxNzI3MDA4ODk2fQ.MLomPHXbmQfWcPchaSv3w9zxsBiezMFgIP7Rf3H9SzU",
  "created_at": "2022-04-09T08:14:11Z",
  "updtaed_at": "2022-04-09T08:14:11Z",
  "user_id": "61614f539f29be942bd9df8e",
  "usercart": [],
  "address": [],
  "orders": []
}
```

- **Add Product (POST)** _[добавление товара (для админов)]_

  http://localhost:8000/admin/addproduct

```json
{
  "product_name": "Смартфон Vivo",
  "price": 23000,
  "rating": 7,
  "image": "abcd.jpg"
}
```

Ответ: "Successfully added"

- **See All Products (GET)** _[получить все товары]_

  http://localhost:8000/users/productview

Ответ:

```json
[
  {
    "ProductID": "66e6d619ed1e10dedc3db0b2",
    "product_name": "Смартфон Vivo",
    "price": 23000,
    "rating": 7,
    "image": "abcd.jpg"
  },
  {
    "ProductID": "66e7313fef58f0b665ef7c7c",
    "product_name": "Iphone",
    "price": 67000,
    "rating": 9,
    "image": "iphone.jpg"
  },
  {
    "ProductID": "66e73167ef58f0b665ef7c7d",
    "product_name": "Штаны",
    "price": 30,
    "rating": 10,
    "image": "pants.jpg"
  },
  {
    "ProductID": "66e7327cef58f0b665ef7c7e",
    "product_name": "Светлое Pivo",
    "price": 120,
    "rating": 10,
    "image": "beer.jpg"
  }
]
```

- **Search Product (GET)** _[поиск по товарам]_

http://localhost:8000/users/search?name=IvO

Ответ:

```json
[
  {
    "ProductID": "66e6d619ed1e10dedc3db0b2",
    "product_name": "Смартфон Vivo",
    "price": 23000,
    "rating": 7,
    "image": "abcd.jpg"
  },
  {
    "ProductID": "66e7327cef58f0b665ef7c7e",
    "product_name": "Светлое Pivo",
    "price": 120,
    "rating": 10,
    "image": "beer.jpg"
  }
]
```

### API-вызовы, доступные при регистрации

- **Add product to cart (GET)** _[добавление товара в корзину]_

  http://localhost:8000/addtocart?productID=xxxxx&userID=xxxxx

- **Remove item from cart (GET)** _[удаление из корзины]_

  http://localhost:8000/removeitem?productID=xxxxx&userID=xxxxx

- **Get user cart (GET)** _[получить корзину и ее общую стоимость]_

  http://localhost:8000/listcart?userID=xxxxx

- **Add delivery address (POST)** _[добавить адрес доставки]_

  http://localhost:8000/addadress?userID=xxxxx

```json
{
  "house_name": "9",
  "street_name": "White street",
  "city_name": "Moscow",
  "post_code": "142321"
}
```

- **Edit home address (PUT)** _[изменить основной адрес]_

  http://localhost:8000/edithomeaddress?userID=xxxxx

- **Edit work address (PUT)** _[изменить запасной адрес]_

  http://localhost:8000/editworkaddress?userID=xxxxx

- **Delete addresses (GET)** _[удалить адреса]_

  http://localhost:8000/deleteaddresses?userID=xxxxx

- **Cart checkout (GET)** _[заказ корзины]_

  http://localhost:8000/cartcheckout?userID=xxxxx

- **Instant buy (GET)** _[купить товар мгновенно]_

  http://localhost:8000/instantbuy?userID=xxxxx&productID=xxxxx

  ![structure](structure.png)

_Проект еще находится в разработке и улучшается..._
