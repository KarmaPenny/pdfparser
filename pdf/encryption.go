package pdf

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
)

var padding_string []byte = []byte("\x28\xBF\x4E\x5E\x4E\x75\x8A\x41\x64\x00\x4E\x56\xFF\xFA\x01\x08\x2E\x2E\x00\xB6\xD0\x68\x3E\x80\x2F\x0C\xA9\xFE\x64\x53\x69\x7A")

func compute_encryption_key(password []byte, trailer Dictionary) ([]byte, error) {
	// pad or truncate password to exactly 32 bytes
	if len(password) < 32 {
		password = append(password, padding_string[:32 - len(password)]...)
	} else {
		password = password[:32]
	}

	// get the encrypt dictionary
	encrypt, err := trailer.GetDictionary("Encrypt")
	if err != nil {
		return nil, err
	}
	Debug("Encrypt = %s", encrypt.String())

	// get O entry from encrypt dictionary
	o_entry, err := encrypt.GetString("O")
	if err != nil {
		return nil, err
	}

	// get U entry from encrypt dictionary
	u_entry, err := encrypt.GetString("U")
	if err != nil {
		return nil, err
	}

	// get P entry from encrypt dictionary
	p_entry, err := encrypt.GetInt("P")
	if err != nil {
		return nil, err
	}
	p_value := make([]byte, 4)
	binary.LittleEndian.PutUint32(p_value, uint32(p_entry))

	// get the id array from trailer dictionary
	ids, err := trailer.GetArray("ID")
	if err != nil {
		return nil, err
	}
	id0, err := ids.GetString(0)
	if err != nil {
		return nil, err
	}

	// get R entry from encrypt dictionary
	revision, err := encrypt.GetInt("R")
	if err != nil {
		return nil, err
	}

	// get EncryptMetaData entry from encrypt dictionary, default true
	encrypt_meta_data, err := encrypt.GetBool("EncryptMetaData")
	if err != nil {
		encrypt_meta_data = true
	}

	// get key length from encrypt dictionary
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

	// create md5sum from all the things
	hash := md5.New()
	hash.Write(password)
	hash.Write([]byte(o_entry))
	hash.Write(p_value)
	hash.Write([]byte(id0))
	if revision >= 4 && !encrypt_meta_data {
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

	// truncate key to correct size
	encryption_key = encryption_key[:key_length]

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
			Debug("%x != %x", []byte(u_entry), u_computed)
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
			Debug("%x != %x", []byte(u_entry)[:16], crypt_sum)
			return encryption_key, NewError("Incorrect Password")
		}
		return encryption_key, nil
	}

	// unsupported revision
	return encryption_key, NewError("Unsupported Revision")
}
