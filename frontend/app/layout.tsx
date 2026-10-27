import type { Metadata } from "next";
import "@fontsource-variable/inter";
import "@fontsource-variable/jetbrains-mono";

import "@/_app/styles/globals.css";
import { AppProviders } from "@/_app/providers/app-providers";

export const metadata: Metadata = {
  title: "ФСП — личный кабинет",
  description: "Личный кабинет Федерации спортивного программирования России",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="ru">
      <body>
        <AppProviders>{children}</AppProviders>
      </body>
    </html>
  );
}
