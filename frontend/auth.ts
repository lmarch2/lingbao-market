import NextAuth from "next-auth"
// import GitHub from "next-auth/providers/github"
import Credentials from "next-auth/providers/credentials"
import { serverApiUrl } from "@/lib/api"

type LoginResponse = {
  token: string;
  username: string;
  id: string;
  isAdmin: boolean;
}
 
export const { handlers, signIn, signOut, auth } = NextAuth({
  providers: [
    // GitHub, // Disabled until configured
    Credentials({
      name: "Lingbao Account",
      credentials: {
        username: { label: "Username", type: "text" },
        password: { label: "Password", type: "password" },
        captchaId: { label: "Captcha ID", type: "text" },
        captchaCode: { label: "Captcha", type: "text" }
      },
      authorize: async (credentials) => {
        if (!credentials?.username || !credentials?.password) return null;

        try {
          const res = await fetch(serverApiUrl("/api/v1/auth/login"), {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              username: credentials.username,
              password: credentials.password,
              captchaId: credentials.captchaId,
              captchaCode: credentials.captchaCode,
            }),
          });

          if (!res.ok) return null;

          const user = (await res.json()) as LoginResponse;
          // User: { token, username, id, isAdmin }
          if (user.token) {
            return {
                id: user.id,
                name: user.username,
                accessToken: user.token,
                isAdmin: user.isAdmin,
            }
          }
          return null
        } catch (e) {
          console.error("Auth error:", e)
          return null
        }
      }
    })
  ],
  callbacks: {
    jwt: async ({ token, user }) => {
      if (user) {
        if (typeof user.accessToken === 'string') {
          token.accessToken = user.accessToken
        }
        token.id = user.id
        token.isAdmin = Boolean(user.isAdmin)
      }
      return token
    },
    session: async ({ session, token }) => {
      if (token) {
        if (typeof token.accessToken === 'string') {
          session.accessToken = token.accessToken
        }
        if (session.user) {
          session.user.isAdmin = Boolean(token.isAdmin)
          if (typeof token.id === 'string' && token.id) {
            session.user.id = token.id
          }
        }
      }
      return session
    },
    authorized: async ({ auth }) => {
      return !!auth
    },
  },
})
