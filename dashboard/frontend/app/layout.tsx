import type { Metadata, Viewport } from 'next';
import './globals.css';
import { InstanceProvider } from '@/components/instance-context';
import { HealthProvider } from '@/components/health-context';
import { AppShell } from '@/components/app-shell';

export const metadata: Metadata = {
  title: {
    default: 'Sloper Console',
    template: '%s · Sloper',
  },
  description: 'Observability and orchestration console for Sloper agent instances.',
};

export const viewport: Viewport = {
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: '#eef6ff' },
    { media: '(prefers-color-scheme: dark)', color: '#0c1020' },
  ],
  colorScheme: 'light dark',
};

const themeScript = `
(function () {
  try {
    var saved = localStorage.getItem('sloper-theme');
    var theme = saved === 'light' || saved === 'dark' ? saved : 'light';
    document.documentElement.dataset.theme = theme;
  } catch (_) {}
})();
`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" data-theme="light" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body>
        <InstanceProvider>
          <HealthProvider>
            <AppShell>{children}</AppShell>
          </HealthProvider>
        </InstanceProvider>
      </body>
    </html>
  );
}
