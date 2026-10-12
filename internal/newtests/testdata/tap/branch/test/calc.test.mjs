import { test } from 'node:test';
import assert from 'node:assert';
import { add, sub } from '../src/calc.mjs';

test('add works', () => {
  assert.strictEqual(add(2, 3), 5);
});

// sub gains a subtest on the branch, five, which fails at the base: its old parent then fails only by it.
test('sub', async (t) => {
  await t.test('zero', () => {
    assert.strictEqual(sub(5, 0), 5);
  });
  await t.test('five', () => {
    assert.strictEqual(sub(5, 3), 2);
  });
});
