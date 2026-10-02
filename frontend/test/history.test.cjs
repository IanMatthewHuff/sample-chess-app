const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const ts = require('typescript');

test('move history renders numbered vertical rows and follows the latest move', () => {
    const source = fs.readFileSync(path.join(__dirname, '..', 'src', 'history.ts'), 'utf8');
    const code = ts.transpileModule(source, {
        compilerOptions: { module: ts.ModuleKind.CommonJS }
    }).outputText;
    const panel = { scrollTop: 0, scrollHeight: 0, clientHeight: 100 };
    const list = {
        dataset: {},
        children: [],
        replaceChildren(...rows) {
            this.children = rows;
            panel.scrollHeight = rows.length * 40;
        }
    };
    const document = {
        getElementById: id => id === 'move-list' ? list : panel,
        createElement: () => ({
            children: [],
            append(...children) { this.children.push(...children); }
        })
    };
    const exports = {};
    vm.runInNewContext(code, { exports, document });
    const moves = ['e4', 'e5', 'Nf3'].map(san => ({ san }));
    exports.renderMoveHistory(moves);
    assert.equal(list.children.length, 2);
    assert.deepEqual(list.children[0].children.map(cell => cell.textContent), ['1.', 'e4', 'e5']);
    assert.deepEqual(list.children[1].children.map(cell => cell.textContent), ['2.', 'Nf3', '']);
    const firstRow = list.children[0];
    exports.renderMoveHistory(moves);
    assert.equal(list.children[0], firstRow);

    panel.scrollHeight = 400;
    panel.scrollTop = 0;
    exports.renderMoveHistory([...moves, { san: 'Nc6' }]);
    assert.equal(panel.scrollTop, 0);

    panel.scrollHeight = 400;
    panel.scrollTop = 300;
    exports.renderMoveHistory([...moves, { san: 'Nc6' }, { san: 'Bb5' }]);
    assert.equal(panel.scrollTop, panel.scrollHeight);
});
