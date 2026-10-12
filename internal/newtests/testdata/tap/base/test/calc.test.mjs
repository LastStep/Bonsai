import { test } from 'node:test';
import assert from 'node:assert';
import { add, sub } from '../src/calc.mjs';

test('add works', () => {
  assert.strictEqual(add(2, 3), 5);
});

test('sub', async (t) => {
  await t.test('zero', () => {
    assert.strictEqual(sub(5, 0), 5);
  });
});
