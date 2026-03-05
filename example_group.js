import { group } from 'k6/x/async';
import { sleep } from 'k6';
import http from 'k6/http';

export default async function () {
  // Example 1: Async group with async operations
  await group('Async API Calls', async () => {
    const response = http.get('https://test.k6.io');
    console.log('Response status:', response.status);
    await sleep(1);
  });

  // Example 2: Nested async groups
  await group('Parent Group', async () => {
    console.log('In parent group');

    await group('Child Group 1', async () => {
      console.log('In child group 1');
      await sleep(0.5);
    });

    await group('Child Group 2', async () => {
      console.log('In child group 2');
      await sleep(0.5);
    });
  });

  // Example 3: Works with sync functions too
  await group('Sync Group', () => {
    console.log('This is a sync function');
    return 'sync result';
  });
}
