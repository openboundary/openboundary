import { betterAuth } from 'better-auth';

export const auth = betterAuth({
  session: {
    strategy: 'jwt',
    expiresIn: 3600,
  },
  emailAndPassword: {
    enabled: true,
  },
  socialProviders: {
    github: {
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    },
  },
});
