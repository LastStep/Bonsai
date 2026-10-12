import { test } from 'node:test';
import assert from 'node:assert';
import { abs } from '../src/neg.mjs';

test('abs works', () => {
  assert.strictEqual(abs(-2), 2);
});
