/* (c) 2019 Confio UO. Licensed under Apache-2.0 */

/* Generated with cbindgen:0.9.1 */

/* Warning, this file is autogenerated by cbindgen. Don't modify this manually. */

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

typedef struct Buffer {
  uint8_t *ptr;
  uintptr_t len;
  uintptr_t cap;
} Buffer;

typedef struct cache_t {

} cache_t;

typedef struct db_t {

} db_t;

typedef struct DB_vtable {
  int64_t (*read_db)(db_t*, Buffer, Buffer);
  void (*write_db)(db_t*, Buffer, Buffer);
} DB_vtable;

typedef struct DB {
  db_t *state;
  DB_vtable vtable;
} DB;

typedef struct api_t {

} api_t;

typedef struct GoApi_vtable {
  int32_t (*humanize_address)(const api_t*, Buffer, Buffer);
  int32_t (*canonicalize_address)(const api_t*, Buffer, Buffer);
} GoApi_vtable;

typedef struct GoApi {
  const api_t *state;
  GoApi_vtable vtable;
} GoApi;

Buffer create(cache_t *cache, Buffer wasm, Buffer *err);

void free_rust(Buffer buf);

Buffer get_code(cache_t *cache, Buffer id, Buffer *err);

Buffer handle(cache_t *cache,
              Buffer code_id,
              Buffer params,
              Buffer msg,
              DB db,
              GoApi api,
              uint64_t gas_limit,
              uint64_t *gas_used,
              Buffer *err);

cache_t *init_cache(Buffer data_dir, uintptr_t cache_size, Buffer *err);

bool init_seed(Buffer pk, Buffer encrypted_key, Buffer *err);

Buffer instantiate(cache_t *cache,
                   Buffer contract_id,
                   Buffer params,
                   Buffer msg,
                   DB db,
                   GoApi api,
                   uint64_t gas_limit,
                   uint64_t *gas_used,
                   Buffer *err);

Buffer key_gen(Buffer *err);

Buffer query(cache_t *cache,
             Buffer code_id,
             Buffer msg,
             DB db,
             GoApi api,
             uint64_t gas_limit,
             uint64_t *gas_used,
             Buffer *err);

/**
 * frees a cache reference
 *
 * # Safety
 *
 * This must be called exactly once for any `*cache_t` returned by `init_cache`
 * and cannot be called on any other pointer.
 */
void release_cache(cache_t *cache);
