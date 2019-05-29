package pdf

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
)

var padding_string []byte = []byte("\x28\xBF\x4E\x5E\x4E\x75\x8A\x41\x64\x00\x4E\x56\xFF\xFA\x01\x08\x2E\x2E\x00\xB6\xD0\x68\x3E\x80\x2F\x0C\xA9\xFE\x64\x53\x69\x7A")

func compute_encryption_key(password []byte, trailer Dictionary) ([]byte, error) {
	// get the encrypt dictionary
	encrypt, err := trailer.GetDictionary("Encrypt")
	if err != nil {
		return nil, err
	}

	// Algorithm 2: Computing an encryption key
	// Step a) pad or truncate password to exactly 32 bytes
	if len(password) < 32 {
		password = append(password, padding_string[:32 - len(password)]...)
	} else {
		password = password[:32]
	}

	// Step b) Initialize md5 hash function and pass result of step a as input
	hash := md5.New()
	hash.Write(password)

	// Step c) Pass the value of encryption dictionary's O entry to the md5 hash function
	o_entry, err := encrypt.GetString("O")
	if err != nil {
		return nil, err
	}
	hash.Write([]byte(o_entry))

	// Step d) convert encrypt dictionary's P entry to little endian uint32 and add to md5 hash function
	p_entry, err := encrypt.GetInt("P")
	if err != nil {
		return nil, err
	}
	p_value := make([]byte, 4)
	binary.LittleEndian.PutUint32(p_value, uint32(p_entry))
	hash.Write(p_value)

	// Step e) pass first element of encrypt dictionary's ID array to md5 hash function
	ids, err := trailer.GetArray("ID")
	if err != nil {
		return nil, err
	}
	id0, err := ids.GetString(0)
	if err != nil {
		return nil, err
	}
	hash.Write([]byte(id0))

	// Step f) for revision 4+, if not encrypting meta data pass 4 bytes with value 0xFFFFFFFF to the md5 hash function
	revision, err := encrypt.GetInt("R")
	if err != nil {
		return nil, err
	}
	encrypt_meta_data, err := encrypt.GetBool("EncryptMetaData")
	if err != nil {
		encrypt_meta_data = true
	}
	if revision >= 4 && !encrypt_meta_data {
		hash.Write([]byte("\xff\xff\xff\xff"))
	}

	// Step g) finish the md5
	encryption_key := hash.Sum(nil)

	// Step h) for revision 3+, re-hash first key_length bytes of encryption key 50 times
	key_length, err := encrypt.GetInt("Length")
	if err != nil {
		key_length = 40
	}
	if revision == 2 {
		key_length = 40
	}
	key_length = key_length / 8
	if key_length < 5 {
		key_length = 5
	} else if key_length > 16 {
		key_length = 16
	}
	if revision >= 3 {
		for i := 0; i < 50; i++ {
			hash = md5.New()
			hash.Write(encryption_key[:key_length])
			encryption_key = hash.Sum(nil)
		}
	}

	// Step i) set encryption key to first key_length bytes of final hash
	encryption_key = encryption_key[:key_length]

	// get U entry from encrypt dictionary
	u_entry, err := encrypt.GetString("U")
	if err != nil {
		return nil, err
	}

	// for revision 3+
	// Algorithm 5: Computing the encryption dictionary's U value
	if revision >= 3 {
		// Step a) Create encryption key using Algorithm 2
		// already done

		// Step b) Initialize md5 has function and pass padding string as input
		hash = md5.New()
		hash.Write(padding_string)

		// Step c) pass first element of encrypt dictionary's ID array to md5 hash and finish the hash
		hash.Write([]byte(id0))
		sum := hash.Sum(nil)

		// Step d & e) rc4 encrypt sum 20 times using key XORed with counter 0-19
		temp_key := make([]byte, len(encryption_key))
		for i := 0; i < 20; i++ {
			for j := range encryption_key {
				temp_key[j] = encryption_key[j] ^ byte(i)
			}
			cipher, err := rc4.NewCipher(temp_key)
			if err != nil {
				return encryption_key, err
			}
			cipher.XORKeyStream(sum, sum)
		}

		// compare first 16 bytes of U entry to sum
		if string([]byte(u_entry)[:16]) != string(sum) {
			return encryption_key, ErrorPassword
		}
		return encryption_key, nil
	}

	// if revision 2 use algorithm 4
	if revision == 2 {
		cipher, err := rc4.NewCipher(encryption_key)
		if err != nil {
			return encryption_key, err
		}

		// calulate u entry
		u_computed := make([]byte, 32)
		cipher.XORKeyStream(u_computed, padding_string)
		if string(u_computed) != u_entry {
			return encryption_key, ErrorPassword
		}
		return encryption_key, nil
	}

	// unsupported revision
	return encryption_key, NewError("Unsupported Revision")
}
