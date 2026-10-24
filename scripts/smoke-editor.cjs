// Local browser smoke. Requires Playwright; never downloads a browser or package.
const assert = require('node:assert/strict');
const { execFileSync, spawn } = require('node:child_process');
const { randomBytes } = require('node:crypto');
const { mkdtempSync, rmSync } = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { once } = require('node:events');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');

async function main() {
  const repo = path.resolve(__dirname, '..');
  const temporary = mkdtempSync(path.join(os.tmpdir(), 'gossip-editor-smoke-'));
  const binary = path.join(temporary, 'gossip');
  const dataFile = path.join(temporary, 'content.json');
  const key = randomBytes(24).toString('hex');
  const base = 'http://127.0.0.1:18123';
  let server;
  let browser;
  async function startServer() {
    server = spawn(binary, ['-listen', '127.0.0.1:18123', '-dev-editor', '-data-file', dataFile], {
      cwd: repo, env: { ...process.env, GOSSIP_DEV_EDITOR_KEY: key }, stdio: ['ignore', 'pipe', 'pipe'],
    });
    let output = '';
    server.stdout.on('data', data => { output += data; });
    server.stderr.on('data', data => { output += data; });
    const startup = new Promise((resolve, reject) => {
      server.once('error', reject);
      server.once('exit', code => reject(new Error(`server exited ${code}: ${output}`)));
      server.stdout.on('data', () => {
        if (output.includes('listening on')) resolve();
      });
    });
    await startup;
  }
  try {
    execFileSync('go', ['build', '-o', binary, './cmd/gossip'], { cwd: repo, stdio: 'pipe' });
    await startServer();
    browser = await chromium.launch({
      headless: true,
      ...(process.env.GOSSIP_CHROME_PATH ? { executablePath: process.env.GOSSIP_CHROME_PATH } : {}),
    });
    const context = await browser.newContext();
    const page = await context.newPage();
    // A colliding local process must not be accepted as our server.
    const edit = await page.goto(`${base}/dev/posts/welcome/edit?key=${key}`);
    assert.equal(edit.status(), 200);
    assert.equal(await page.locator('input[name=dev_key]').inputValue(), key);
    const title = 'ローカル編集テスト — Gossip';
    const content = '日本語の本文\n<script>window.gossipInjected = true</script> & <b>text</b>';
    await page.getByLabel('Title').fill(title);
    await page.getByLabel('Content').fill(content);
    const saved = page.waitForResponse(response => response.request().method() === 'POST');
    await page.getByRole('button', { name: 'Save', exact: true }).click();
    assert.equal((await saved).status(), 303);
    await page.waitForURL(`${base}/posts/welcome`);
    assert.equal(await page.locator('h1').textContent(), title);
    assert.equal(await page.locator('.content').textContent(), content);
    assert.equal(await page.locator('.content script, .content b').count(), 0);
    assert.equal(await page.evaluate(() => window.gossipInjected), undefined);
    assert.match(await page.locator('footer').textContent(), /native Go sample plugin/);
    await page.reload();
    assert.equal(await page.locator('.content').textContent(), content);

    assert.equal((await context.request.get(`${base}/dev/posts/welcome/edit`)).status(), 401);
    assert.equal((await context.request.post(`${base}/dev/posts/welcome/edit`, {
      form: { title: 'unauthorized change', content: 'must not save' },
    })).status(), 401);
    await page.reload();
    assert.equal(await page.locator('h1').textContent(), title);

    await page.goto(`${base}/dev/posts/welcome/edit?key=${key}`);
    await page.getByLabel('Title').fill('');
    await page.getByRole('button', { name: 'Save', exact: true }).click();
    assert.equal(await page.getByLabel('Title').evaluate(input => input.validity.valueMissing), true);
    assert.equal((await context.request.post(`${base}/dev/posts/welcome/edit`, {
      form: { dev_key: key, title: '', content: 'must not save' },
    })).status(), 422);
    await page.goto(`${base}/posts/welcome`);
    assert.equal(await page.locator('.content').textContent(), content);
    // Simulate process loss after the save response; no graceful flush is used.
    const stopped = once(server, 'exit');
    server.kill('SIGKILL');
    await stopped;
    await startServer();
    await page.goto(`${base}/posts/welcome`);
    assert.equal(await page.locator('h1').textContent(), title);
    assert.equal(await page.locator('.content').textContent(), content);
    assert.equal(await page.locator('.content script, .content b').count(), 0);
    await page.goto(`${base}/dev/posts/welcome/edit?key=${key}`);
    assert.equal(await page.getByLabel('Title').inputValue(), title);
    assert.equal(await page.getByLabel('Content').inputValue(), content);
    console.log(JSON.stringify({
      status: 'PASS', browser: browser.version(),
      checks: ['native form submit / 303 / view', 'Japanese content', 'HTML escaped',
        'sample plugin footer', 'reload within process', 'missing key rejected',
        'empty title browser and server validation', 'rejected writes leave content unchanged',
        'SIGKILL / restart / view and editor retain saved content'],
      persistence: 'local JSON snapshot; process restart verified, power-loss durability unverified',
    }, null, 2));
  } finally {
    try {
      if (browser) await browser.close();
    } finally {
      if (server && server.exitCode === null && server.signalCode === null) {
        const exited = once(server, 'exit');
        server.kill('SIGTERM');
        await exited;
      }
      rmSync(temporary, { recursive: true, force: true });
    }
  }
}

main().catch(error => { console.error(error); process.exitCode = 1; });
