import {getRequestConfig} from 'next-intl/server';
 
export default getRequestConfig(async ({requestLocale}) => {
  // This typically comes from a routing library or middleware
  let locale = await requestLocale;
 
  // Ensure that a valid locale is used
  if (!locale || !['en'].includes(locale)) {
    locale = 'en';
  }
 
  return {
    locale,
    messages: (await import(`../messages/${locale}.json`)).default
  };
});
