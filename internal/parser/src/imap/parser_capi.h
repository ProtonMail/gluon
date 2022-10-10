#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>

struct IMAPParser;
typedef struct IMAPParser IMAPParser;

IMAPParser* IMAPParser_new();

void IMAPParser_free(IMAPParser* parser);

/// Parser an IMAP input, with a given delimiter. Returns 0 on success -1 on error. See `IMAPParser_getError()` for
/// more details.
int IMAPParser_parse(IMAPParser* parser, const char* input, char delimiter);


/// Return last recoded IMAP tag after a call to `IMAPParser_parse`, may or may not be available.
const char* IMAPParser_getTag(IMAPParser* parser);

/// Return the last recoded error after a call to `IMAPParser_parse` or null if no error took place.
const char* IMAPParser_getError(IMAPParser* parser);

/// Return the last recorded command after a call to `IMAPParser_parse` or null on error.
const void* IMAPParser_getCommandData(IMAPParser* parser);

/// Return the size of recorded command after a call to `IMAPParser_parse` or 0 on error.
int IMAPParser_getCommandSize(IMAPParser* parser);


#ifdef __cplusplus
}
#endif
