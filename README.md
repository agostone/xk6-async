# xk6-async

A k6 extension that provides async-compatible `group()` and `check()` functions.

## Build

```bash
xk6 build --with github.com/agostone/xk6-async
```

## Usage

### group

```javascript
import { group } from 'k6/x/async';
import http from 'k6/http';

export default async function () {
  await group('My Async Group', async () => {
    const response = await http.asyncRequest('GET', 'https://test.k6.io');
    console.log('Response status:', response.status);
  });
}
```

### check

```javascript
import { check } from 'k6/x/async';
import http from 'k6/http';

export default async function () {
  const response = await http.asyncRequest('GET', 'https://test.k6.io');
  
  await check(response, {
    'status is 200': (r) => r.status === 200,
    'async validation': async (r) => {
      // Perform async validation
      return r.status < 400;
    }
  });
}
```

## Features

### group
- ✅ Supports async functions
- ✅ Correctly measures group duration including async operations
- ✅ Properly manages group tags across async boundaries
- ✅ Emits `group_duration` metrics
- ✅ Works with both sync and async callbacks

### check
- ✅ Supports async functions in checks
- ✅ Handles mixed sync and async checks
- ✅ Emits `checks` metrics after async operations complete
- ✅ Works with custom tags
- ✅ Compatible with all check features from built-in `check()`

## Difference from built-in functions

### group vs group()

The built-in `group()` function doesn't support async functions. This extension provides an async-compatible `group()` which:
- Accepts async functions as callbacks
- Waits for Promises to resolve before emitting metrics
- Ensures group tags are correctly restored after async operations complete

### check vs check()

The built-in `check()` function doesn't support async functions. This extension provides an async-compatible `check()` which:
- Accepts async functions in check definitions
- Waits for all Promises to resolve before emitting check metrics
- Supports mixed sync and async checks in the same call
