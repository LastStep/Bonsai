import { test } from 'node:test';
import assert from 'node:assert';
import { sub } from '../src/calc.mjs';

// New on the branch; fails at the base, where sub adds.
test('sub more', () => {
  assert.strictEqual(sub(5, 3), 2);
});
