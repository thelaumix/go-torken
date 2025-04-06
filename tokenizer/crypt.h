#include <stddef.h>
#include <stdint.h>

int CRP_Encrypt_EX(int algo,
     const uint8_t* in,  size_t inLen,
     const uint8_t* key, size_t keyLen,
     const uint8_t* nonce, size_t nonceLen,
     uint8_t* out);
int CRP_Decrypt_EX(int algo,
     const uint8_t* in,  size_t inLen,
     const uint8_t* key, size_t keyLen,
     const uint8_t* nonce, size_t nonceLen,
     uint8_t* out);
