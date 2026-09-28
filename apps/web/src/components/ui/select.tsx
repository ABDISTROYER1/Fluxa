import * as React from 'react';
import { cn } from '@/lib/utils';

export const Select = React.forwardRef<
  HTMLSelectElement,
  React.SelectHTMLAttributes<HTMLSelectElement>
>(({ className, children, ...props }, ref) => {
  return (
    <select
      className={cn(
        'flex h-10 w-full appearance-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-foreground shadow-sm transition-colors focus:border-primary focus:ring-4 focus:ring-primary/10 focus:outline-none disabled:opacity-60 disabled:cursor-not-allowed',
        className,
      )}
      ref={ref}
      {...props}
    >
      {children}
    </select>
  );
});
Select.displayName = 'Select';
