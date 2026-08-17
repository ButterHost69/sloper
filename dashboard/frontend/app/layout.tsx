import type { Metadata, Viewport } from 'next';
import './globals.css';
import { InstanceProvider } from '@/components/instance-context';
import { AppShell } from '@/components/app-shell';

export const metadata: Metadata = {
  title: 'Sloper Console',
  description: 'Observability dashboard for sloper agent instances.',
};

export const viewport: Viewport = {
  themeColor: '#09090b',
  colorScheme: 'dark',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="dark">
      <body className="bg-base text-ink antialiased">
        <InstanceProvider>
          <AppShell>{children}</AppShell>
        </InstanceProvider>
      </body>
    </html>
  );
}