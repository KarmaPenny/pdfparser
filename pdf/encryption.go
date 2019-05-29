package pdf

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
)

var padding_string []byte = []byte("\x28\xbf\x4e\x5e\x4e\x75\x8a\x41\x64\x00\x4e\x56\xff\xfa\x01\x08\x2e\x2e\x00\xb6\x3e\x80\x2f\x0c\xa9\xfe\x64\x53\x69\x7a")

func compute_encryption_key(password []byte, trailer Dictionary) ([]byte, error) {
	// pad or truncate password to exactly 32 bytes
	if len(password) < 32 {
		password = append(password, padding_string[:32 - len(password)]...)
	} else {
		password = password[:32]
	}

	// get the encrypt dictionary
	encrypt, err := trailer.GetDictionary("/Encrypt")
	if err != nil {
		return nil, err
	}

	// get O entry from encrypt dictionary
	o_entry, err := encrypt.GetString("/O")
	if err != nil {
		return nil, err
	}

	// get U entry from encrypt dictionary
	u_entry, err := encrypt.GetString("/U")
	if err != nil {
		return nil, err
	}

	// get P entry from encrypt dictionary
	p_entry, err := encrypt.GetInt("/P")
	if err != nil {
		return nil, err
	}
	p_value := make([]byte, 4)
	binary.LittleEndian.PutUint32(p_value, uint32(p_entry))

	// get the id array from trailer dictionary
	ids, err := trailer.GetArray("/ID")
	if err != nil {
		return nil, err
	}
	id0, err := ids.GetString(0)
	if err != nil {
		return nil, err
	}

	// get R entry from encrypt dictionary
	revision, err := encrypt.GetInt("/R")
	if err != nil {
		return nil, err
	}

	// get EncryptMetaData entry from encrypt dictionary
	encrypt_meta_data := true
	encrypt_meta_data_object, err := encrypt.GetObject("/EncryptMetaData")
	if err == nil {
		if encrypt_meta_data_object.String() == "false" {
			encrypt_meta_data = false
		}
	}

	// get key length from encrypt dictionary
	key_length, err := encrypt.GetInt("/Length")
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

	// create md5sum from all the things
	hash := md5.New()
	hash.Write([]byte(o_entry))
	hash.Write(p_value)
	hash.Write([]byte(id0))
	if revision == 4 && !encrypt_meta_data {
		hash.Write([]byte("\xff\xff\xff\xff"))
	}
	encryption_key := hash.Sum(nil)

	// wash the md5 50 times if revision 3 or greater
	if revision >= 3 {
		for i := 0; i < 50; i++ {
			hash = md5.New()
			encryption_key = hash.Sum(encryption_key[:key_length])
		}
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
			return encryption_key, NewError("Incorrect Password")
		}
		return encryption_key, nil
	}

	// if revision 3+ use algorithm 5
	if revision > 2 {
		hash = md5.New()
		hash.Write(padding_string)
		hash.Write([]byte(id0))
		sum := hash.Sum(nil)
		cipher, err := rc4.NewCipher(encryption_key)
		if err != nil {
			return encryption_key, err
		}
		crypt_sum := make([]byte, 16)
		cipher.XORKeyStream(crypt_sum, sum)
		temp_key := make([]byte, len(encryption_key))
		for i := 1; i < 20; i++ {
			for j := range encryption_key {
				temp_key[j] = encryption_key[j] ^ byte(i)
			}
			cipher, err = rc4.NewCipher(temp_key)
			if err != nil {
				return encryption_key, err
			}
			cipher.XORKeyStream(crypt_sum, crypt_sum)
		}
		if string([]byte(u_entry)[:16]) != string(crypt_sum) {
			return encryption_key, NewError("Incorrect Password")
		}
		return encryption_key, nil
	}

	// unsupported revision
	return encryption_key, NewError("Unsupported Revision")
}
