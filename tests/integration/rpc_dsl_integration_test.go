package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Product struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

// 全局变量跟踪当前产品数量
var currentProductCount int

func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		// First check if the server is listening
		resp, err := http.Get(url + "/api/health")
		if err == nil {
			resp.Body.Close()
			// Then check if RPC endpoint is available
			resp, err = http.Post(url+"/rpc/products/list", "application/json", bytes.NewBuffer([]byte("{}")))
			if err == nil {
				resp.Body.Close()
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("server not ready after %v", timeout)
}

func TestRPCAndDSLIntegration(t *testing.T) {
	// 设置测试超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 使用固定端口 8099
	baseURL := "http://127.0.0.1:8099"

	// 等待服务端口可用
	require.NoError(t, waitForServer(baseURL, 30*time.Second))

	// 重置全局产品计数器
	currentProductCount = 0

	testProduct := Product{
		Name:        "Test Product",
		Price:       100,
		Description: "A test product for integration testing",
	}

	t.Run("Clean All Products", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		reqBody, err := json.Marshal(map[string]interface{}{})
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/clean", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 验证所有产品已被删除
		resp, err = http.Post(baseURL+"/rpc/products/list", "application/json", bytes.NewBuffer([]byte("{}")))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var products []Product
		err = json.NewDecoder(resp.Body).Decode(&products)
		require.NoError(t, err)
		assert.Len(t, products, 0)
		currentProductCount = 0
	})

	t.Run("Create Product", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		reqBody, err := json.Marshal(testProduct)
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/create", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var createdProduct Product
		err = json.NewDecoder(resp.Body).Decode(&createdProduct)
		require.NoError(t, err)
		assert.NotEmpty(t, createdProduct.ID)
		assert.Equal(t, testProduct.Name, createdProduct.Name)
		assert.Equal(t, testProduct.Price, createdProduct.Price)
		assert.Equal(t, testProduct.Description, createdProduct.Description)
		assert.NotEmpty(t, createdProduct.Created)
		testProduct.ID = createdProduct.ID
		currentProductCount++
	})

	t.Run("Get Product", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		// 先创建一个产品
		reqBody, err := json.Marshal(testProduct)
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/create", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var createdProduct Product
		err = json.NewDecoder(resp.Body).Decode(&createdProduct)
		require.NoError(t, err)
		testProduct.ID = createdProduct.ID
		currentProductCount++

		resp, err = http.Get(baseURL + "/rpc/products/product/" + testProduct.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var retrievedProduct Product
		err = json.NewDecoder(resp.Body).Decode(&retrievedProduct)
		require.NoError(t, err)
		assert.Equal(t, testProduct.ID, retrievedProduct.ID)
		assert.Equal(t, testProduct.Name, retrievedProduct.Name)
		assert.Equal(t, testProduct.Price, retrievedProduct.Price)
		assert.Equal(t, testProduct.Description, retrievedProduct.Description)
	})

	t.Run("List Products", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		// 先创建一个产品
		reqBody, err := json.Marshal(testProduct)
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/create", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var createdProduct Product
		err = json.NewDecoder(resp.Body).Decode(&createdProduct)
		require.NoError(t, err)
		testProduct.ID = createdProduct.ID
		currentProductCount++

		reqBody, err = json.Marshal(map[string]interface{}{})
		require.NoError(t, err)
		resp, err = http.Post(baseURL+"/rpc/products/list", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var products []Product
		err = json.NewDecoder(resp.Body).Decode(&products)
		require.NoError(t, err)
		assert.Len(t, products, currentProductCount)
		// 检查最新的产品是否在列表中
		found := false
		for _, product := range products {
			if product.ID == testProduct.ID {
				assert.Equal(t, testProduct.Name, product.Name)
				found = true
				break
			}
		}
		assert.True(t, found, "Created product should be in the list")
	})

	t.Run("Update Product", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		// 先创建一个产品
		reqBody, err := json.Marshal(testProduct)
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/create", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var createdProduct Product
		err = json.NewDecoder(resp.Body).Decode(&createdProduct)
		require.NoError(t, err)
		testProduct.ID = createdProduct.ID
		currentProductCount++

		updateData := map[string]interface{}{
			"id":          testProduct.ID,
			"name":        "Updated Product",
			"price":       200,
			"description": "An updated test product",
		}
		reqBody, err = json.Marshal(updateData)
		require.NoError(t, err)
		resp, err = http.Post(baseURL+"/rpc/products/update", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var updatedProduct Product
		err = json.NewDecoder(resp.Body).Decode(&updatedProduct)
		require.NoError(t, err)
		assert.Equal(t, testProduct.ID, updatedProduct.ID)
		assert.Equal(t, updateData["name"], updatedProduct.Name)
		assert.Equal(t, updateData["price"], updatedProduct.Price)
		assert.Equal(t, updateData["description"], updatedProduct.Description)
	})

	t.Run("Delete Product", func(t *testing.T) {
		select {
		case <-ctx.Done():
			t.Fatal("Test timeout")
		default:
		}

		// 先创建一个产品
		reqBody, err := json.Marshal(testProduct)
		require.NoError(t, err)
		resp, err := http.Post(baseURL+"/rpc/products/create", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		var createdProduct Product
		err = json.NewDecoder(resp.Body).Decode(&createdProduct)
		require.NoError(t, err)
		testProduct.ID = createdProduct.ID
		currentProductCount++

		deleteData := map[string]interface{}{
			"id": testProduct.ID,
		}
		reqBody, err = json.Marshal(deleteData)
		require.NoError(t, err)
		resp, err = http.Post(baseURL+"/rpc/products/delete", "application/json", bytes.NewBuffer(reqBody))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		currentProductCount--

		resp, err = http.Get(baseURL + "/rpc/products/product/" + testProduct.ID)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
