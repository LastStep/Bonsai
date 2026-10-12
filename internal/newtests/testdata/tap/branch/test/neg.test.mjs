import { test } from 'node:test';
import assert from 'node:assert';
import { neg } from '../src/neg.mjs';

// New on the branch; imports what the branch adds, so its file does not load at the base.
test('neg works', () => {
  assert.strictEqual(neg(2), -2);
});
