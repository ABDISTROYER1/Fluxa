'use client';

import { X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Button } from './button';
import { Card } from './card';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export function Modal({ open, onClose, title, description, children, className }: ModalProps) {
  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 animate-in fade-in duration-200"
      role="dialog"
      aria-modal="true"
    >
      <div className="absolute inset-0 bg-black/40 backdrop-blur-sm" onClick={onClose} />
      <Card
        className={cn(
          'relative z-10 w-full max-w-lg shadow-xl animate-in zoom-in-95 duration-200',
          className,
        )}
      >
        <div className="flex items-start justify-between gap-4 p-6 pb-0">
          <div className="flex flex-col gap-1">
            <h2 className="text-lg font-semibold text-foreground">{title}</h2>
            {description && <p className="text-sm text-muted-foreground">{description}</p>}
          </div>
          <Button
            variant="ghost"
            size="sm"
            className="h-8 w-8 p-0"
            onClick={onClose}
            aria-label="Close"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
        <div className="p-6">{children}</div>
      </Card>
    </div>
  );
}
