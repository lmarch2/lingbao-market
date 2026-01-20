import createMiddleware from 'next-intl/middleware';
 
const intlMiddleware = createMiddleware({
  locales: ['en'],
  defaultLocale: 'en'
});

export default intlMiddleware;
 
export const config = {
  // Match only internationalized pathnames
  matcher: ['/', '/en/:path*']
};
