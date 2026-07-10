import type { User } from '../users/types';

export type Session = {
  id: string;
  ip: string;
  browser: string;
  creation_time: string;
  user_id: number;
};

export type UserProfileResponse = {
  user: User;
  sessions: Session[];
};

export type ChangePasswordPayload = {
  old_password: string;
  new_password: string;
};
