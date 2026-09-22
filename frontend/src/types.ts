export interface PollOption {
  id: string
  label: string
  voteCount: number
}

export interface Poll {
  id: string
  title: string
  options: PollOption[]
  createdAt: string
  userVotedOptionId?: string
  createdByUser: boolean
}

