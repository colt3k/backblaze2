# backblaze2

map to api provided by back blaze b2 cloud

## Testing failures on upload

X-Bz-Test-Mode: fail_some_uploads
- Add this to b2_upload_file API call, 
  - causes intermittent artificial failures, allowing you to test the resiliency of your code.

## Testing authorization failures
Add this to authorization header
- X-Bz-Test-Mode: expire_some_account_authorization_tokens

## Testing Upload 403 Forbidden error
X-Bz-Test-Mode: force_cap_exceeded 
- header before making upload-related API calls. 
  - This will cause a cap limit failure, allowing you to verify correct behavior of your code.
    
    
## Retry Header
429 Too Many Requests. 
- B2 may limit API requests on a per-account basis. When the 429 status code is returned 
  - from an API, the response will also include a "Retry-After" header where the value is the number of seconds a 
  - developer should wait before re-issuing the command. This status code may be returned on any B2 API.    
