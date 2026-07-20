package shortener

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
)

var ErrDuplicateShortURL = errors.New("Dulplicate url")
// to generate sha256....

// func sha256C(data []byte) [size]byte {

// 	h := sha256.New() // to intialize sha256...

// 	h.Write(data)   	// Write the data into it

// 	var sum [size]byte    //  initialize fixed-size array


// 	h.Sum(sum[:0])   // Extract the final hash into your array

// 	return sum
// }


// or

func generateHash(data []byte) [32]byte {
    return sha256.Sum256(data)
}

const alpha = "01234567889abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ" 
const base = 62


func base62Gen (hashed []byte) string {

	n := new(big.Int).SetBytes(hashed)  // big is is basically a struct that contian chunks of slice..

	if n.Sign() == 0 {
		return string(alpha[0])
	}

	var result []byte  // we will store result here...

	bigBase := big.NewInt(base)  // convert base into *big.Int struct ..
	zero := big.NewInt(0) 
	temp := new(big.Int).Set(n) // allocate new var
	rem := new(big.Int)  


	for temp.Cmp(zero) > 0 {
		temp.DivMod(temp,bigBase,rem)  // 
		result = append(result, alpha[rem.Int64()])
	}


	for i, j := 0, len(result) - 1; i < j; i, j = i + 1, j - 1 {
		result[i],result[j] = result[j],result[i]
	}
	return string(result)
}

const maxRetries = 10

func GenerateShortURL (originalURL string, repo URLRepository) (string,error) {
	for attempt := 0;attempt < maxRetries; attempt++ {
		targetURL := originalURL
		if attempt > 0 {
			targetURL = fmt.Sprintf("%s:collisioon-salt:%d",originalURL,attempt)
		}
		hash := generateHash([]byte(targetURL))
		encode := base62Gen(hash[:])  // to convert arr --> slice
		for len(encode) < 7 {
			encode = "0" + encode
		}
		shortURL := encode[:7]

		// insert into db...
		err := repo.Save(URLMapping{
			LongURL: originalURL,
			ShortURL: shortURL,
		})
		
		if err == nil {
			// insert succeced, no collision........
			return shortURL, nil
		}

		// insertion failed (duplicate or real collision)
		existing, findErr := repo.FindByShortURL(shortURL) 
		if findErr != nil {
			return "",findErr
		}
		if existing.LongURL == originalURL {
			//same url (return existing shortyy) 
			return existing.ShortURL, nil
		}
	}
	return "",fmt.Errorf("Could not genrate unique short URL after %d attempt",maxRetries)
	
}