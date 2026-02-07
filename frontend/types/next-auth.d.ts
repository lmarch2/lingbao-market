import { DefaultSession } from 'next-auth';

declare module 'next-auth' {
  interface Session {
    accessToken?: string;
    user?: {
      id?: string;
      isAdmin?: boolean;
    } & DefaultSession['user'];
  }

  interface User {
    accessToken?: string;
    isAdmin?: boolean;
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    accessToken?: string;
    id?: string;
    isAdmin?: boolean;
  }
}
