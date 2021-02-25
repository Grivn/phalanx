#ifndef USIG_H__
#define USIG_H__

#include <stddef.h>
#include <stdint.h>

#include <sgx_urts.h>
#include <sgx_tcrypto.h>

#ifdef __cplusplus
extern "C" {
#endif

/**
 * usig_init() - Create and initialize an instance of USIG enclave
 * @enclave_file:     path to the enclave image file
 * @enclave_id:       pointer to store enclave ID to
 * @sealed_data:      pointer to sealed key or NULL
 * @selaed_data_size: length of the buffer @sealed_data points to
 *
 * The sealed key can be obtained by usig_seal_key() from another USIG
 * instance created on the same machine. If non-NULL pointer is
 * supplied as @sealed_data then the enclave will use that data to
 * unseal the key pair to produce signatures, otherwise a new key pair
 * will be generated. In any case, each enclave instance will
 * initialize its epoch value randomly.
 *
 * The normal sequence would be to create the first USIG instance by
 * passing NULL as @sealed_data. A new key pair will be generated by
 * the enclave in that case. Then this key pair can be sealed using
 * usig_seal_key(). The sealed key pair can be stored permanently and
 * reused on subsequent invocations to usig_init().
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_init(const char *enclave_file, sgx_enclave_id_t *enclave_id,
                       void *sealed_data, size_t sealed_data_size);

/**
 * usig_destroy() - Destroy a USIG enclave instance
 * @enclave_id: enclave ID returned by successful usig_init()
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_destroy(const sgx_enclave_id_t enclave_id);

/**
 * usig_create_ui() - Assign the next USIG counter value to a digest
 * @enclave_id: enclave ID of the USIG instance returned by usig_init()
 * @digest:     message digest to create UI for
 * @counter:    pointer to store the assigned counter value to
 * @signature:  pointer to store USIG signature to
 *
 * This invocation will increment the value of the ephemeral counter
 * maintained by the enclave instance and create a signature. The
 * signature covers the message digest followed by the epoch and
 * counter values in little-endian byte order. The signature is
 * produced using the key pair of the enclave instance.
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_create_ui(sgx_enclave_id_t enclave_id,
                            sgx_sha256_hash_t digest,
                            uint64_t *counter,
                            sgx_ec256_signature_t *signature);

/**
 * usig_get_epoch() - Get epoch value of the USIG instance
 * @enclave_id: enclave ID of the USIG instance returned by
 *              usig_init()
 * @epoch:      pointer to store the epoch value to
 *
 * Epoch is a unique value for each new USIG instance. It is randomly
 * generated by the enclave during initialization.
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_get_epoch(sgx_enclave_id_t enclave_id,
                            uint64_t *epoch);

/**
 * usig_get_pub_key() - Get public key of the USIG instance
 * @enclave_id: enclave ID of the USIG instance returned by
 *              usig_init()
 * @pub_key:    pointer to store the public key to
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_get_pub_key(sgx_enclave_id_t enclave_id,
                              sgx_ec256_public_t *pub_key);

/**
 * usig_seal_key() - Get sealed key of the USIG instance
 * @enclave_id:       enclave ID of the USIG instance returned by
 *                    usig_init()
 * @sealed_data:      pointer to store the address to the sealed data
 *                    to. It is a responsibility of the caller to free
 *                    the memory buffer allocated by successful
 *                    invocation of this function
 * @sealed_data_size: pointer to store sealed data size to
 *
 * The retrieved sealed key can be supplied to usig_init(). The sealed
 * key is only valid on the same hardware platform, i.e. it cannot be
 * transferred to and reused on another physical machine.
 *
 * Return: SGX_SUCCESS if no error; SGX error status, otherwise
 */
sgx_status_t usig_seal_key(sgx_enclave_id_t enclave_id,
                           void **sealed_data,
                           size_t *sealed_data_size);

#ifdef __cplusplus
}
#endif

#endif // USIG_H__
