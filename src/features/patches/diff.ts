export type DiffKind = 'same' | 'add' | 'remove';
export interface DiffLine { kind: DiffKind; text: string }

export function lineDiff(before: string, after: string): DiffLine[] {
  const left = before.split('\n');
  const right = after.split('\n');
  const table = Array.from({ length: left.length + 1 }, () => new Uint16Array(right.length + 1));
  for (let i = left.length - 1; i >= 0; i--) {
    for (let j = right.length - 1; j >= 0; j--) {
      table[i][j] = left[i] === right[j] ? table[i + 1][j + 1] + 1 : Math.max(table[i + 1][j], table[i][j + 1]);
    }
  }
  const result: DiffLine[] = [];
  let i = 0;
  let j = 0;
  while (i < left.length || j < right.length) {
    if (i < left.length && j < right.length && left[i] === right[j]) {
      result.push({ kind: 'same', text: left[i++] });
      j++;
    } else if (j < right.length && (i === left.length || table[i][j + 1] >= table[i + 1][j])) {
      result.push({ kind: 'add', text: right[j++] });
    } else {
      result.push({ kind: 'remove', text: left[i++] });
    }
  }
  return result;
}
