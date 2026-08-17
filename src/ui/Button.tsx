import { cn } from '../lib/cn';

interface ButtonProps {
  children: unknown;
  disabled?: boolean;
  onClick?: () => void;
  type?: 'button' | 'submit';
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md';
  className?: string;
}

const variants = {
  primary: 'bg-lime-300 font-bold text-black disabled:bg-white/5 disabled:text-zinc-600',
  secondary: 'bg-zinc-100 font-bold text-black disabled:opacity-30',
  ghost: 'border border-white/10 text-zinc-200 hover:bg-white/5 disabled:opacity-30',
};

const sizes = { sm: 'px-3 py-2 text-xs', md: 'px-4 py-2.5 text-xs' };

export function Button(props: ButtonProps) {
  return <button type={props.type ?? 'button'} disabled={props.disabled} onClick={props.onClick} className={cn('rounded-lg transition-colors', variants[props.variant ?? 'ghost'], sizes[props.size ?? 'sm'], props.className)}>{props.children}</button>;
}
