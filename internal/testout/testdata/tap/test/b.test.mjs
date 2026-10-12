// A fixture of package testout's tests (internal/testout/testdata/tap): node's test runner prints it as TAP in out.tap,
// and the tests run single tests of it by the commands package testout gives.
import { describe, it, test } from 'node:test';
import assert from 'node:assert';

describe('outer (suite)', () => {
  it('passes', () => {});
  describe('inner [x]', () => {
    it('fails: a.b + c?', () => { assert.strictEqual(1, 2); });
    it('unrelated', () => { assert.ok(false, 'unrelated failure'); });
  });
});

test('second top', async (t) => {
  await t.test('sub a', () => {});
  await t.test('sub b', async (t) => {
    await t.test('sub b deep', () => { assert.ok(false, 'deep failed'); });
  });
});

describe('mixed', () => {
  it('holder', async (t) => {
    await t.test('child', () => { assert.ok(false, 'child failed'); });
  });
});

test('other', () => { assert.ok(false, 'other failed'); });
test('a failing todo', { todo: true }, () => { assert.ok(false, 'not yet'); });
test('bare skip', (t) => { t.skip(); });
