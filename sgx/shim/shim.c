#include <stdlib.h>
#include <stdbool.h>

#include "usig.h"
#include "usig_u.h"

// Wraps an ECall invocation assuming the convention of 'ecall_usig_'
// prefix of the function name and sgx_status_t its return value type
#define ECALL_USIG(enclave_id, fn, ...)                         \
        ({                                                      \
                sgx_status_t sgx_ret, ret;                      \
                sgx_ret = ecall_usig_##fn(enclave_id, &ret,     \
                                          ##__VA_ARGS__);       \
                sgx_ret != SGX_SUCCESS ? sgx_ret : ret;         \
        })

static sgx_launch_token_t token;

sgx_status_t usig_init(const char *enclave_file, sgx_enclave_id_t *enclave_id,
                       void *sealed_data, size_t sealed_data_size)
{
        sgx_status_t ret;
        int updated = 0;

        ret = sgx_create_enclave(enclave_file, SGX_DEBUG_FLAG, &token,
                                 &updated, enclave_id, NULL);
        if (ret != SGX_SUCCESS) {
                goto err_out;
        }
        ret = ECALL_USIG(*enclave_id, init, sealed_data, sealed_data_size);
        if (ret != SGX_SUCCESS) {
                goto err_enclave_created;
        }

        return SGX_SUCCESS;

err_enclave_created:
        sgx_destroy_enclave(*enclave_id);
err_out:
        return ret;
}

sgx_status_t usig_destroy(const sgx_enclave_id_t enclave_id)
{
        return sgx_destroy_enclave(enclave_id);
}

sgx_status_t usig_create_ui(sgx_enclave_id_t enclave_id,
                            sgx_sha256_hash_t digest,
                            uint64_t *counter,
                            sgx_ec256_signature_t *signature)
{
        return ECALL_USIG(enclave_id, create_ui, digest, counter, signature);
}

sgx_status_t usig_get_epoch(sgx_enclave_id_t enclave_id,
                            uint64_t *epoch)
{
        return ECALL_USIG(enclave_id, get_epoch, epoch);
}

sgx_status_t usig_get_pub_key(sgx_enclave_id_t enclave_id,
                              sgx_ec256_public_t *pub_key)
{
        return ECALL_USIG(enclave_id, get_pub_key, pub_key);
}

sgx_status_t usig_seal_key(sgx_enclave_id_t enclave_id,
                           void **sealed_data,
                           size_t *sealed_data_size)
{
        sgx_status_t ret;
        uint32_t size;
        void *buf;

        ret = ECALL_USIG(enclave_id, get_sealed_key_size, &size);
        if (ret != SGX_SUCCESS) {
                goto err_out;
        }
        *sealed_data_size = size;

        buf = malloc(size);
        if (!buf) {
                // Exhausted virtual address space is an unrecoverable
                // error that should never happen
                abort();
        }

        ret = ECALL_USIG(enclave_id, seal_key, buf, size);
        if (ret != SGX_SUCCESS) {
                goto err_buf_allocated;
        }
        *sealed_data = buf;

        return SGX_SUCCESS;

err_buf_allocated:
        free(buf);
err_out:
        return ret;
}