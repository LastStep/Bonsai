import { test } from 'node:test';
import assert from 'node:assert';
import { mul } from '../src/mul.mjs';

// New on the branch; passes at the base too: it proves nothing new.
test('mul works', () => {
  assert.strictEqual(mul(2, 3), 6);
});
