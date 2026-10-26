async function getExactWorldRecords() {
  const gameId = "blinx1" // Change to your game slug/ID

  console.log(`Fetching game structure for ${gameId}...`)

  // 1. Fetch game details with embedded levels and categories
  const gameRes = await fetch(
    `/api/v1/games/${gameId}?embed=levels,categories`,
  )
  const gameData = await gameRes.json()

  if (!gameData.data) {
    console.error("Game not found!")
    return
  }

  const levels = gameData.data.levels.data || []
  const categories = gameData.data.categories.data || []

  // Filter for individual level categories (types that apply to levels, not full-game)
  // On speedrun.com, IL categories usually have "per-level" type
  const ilCategories = categories.filter(
    (cat) => cat.type === "per-level",
  )

  console.log(
    `Found ${levels.length} levels and ${ilCategories.length} IL categories.`,
  )
  console.log(
    `Total expected WR combinations: ~${levels.length * ilCategories.length}`,
  )

  const worldRecords = []

  // 2. Loop through each level and IL category to fetch the leaderboard
  for (const level of levels) {
    for (const cat of ilCategories) {
      try {
        // Fetch leaderboard for this specific level and category
        const lbRes = await fetch(
          `/api/v1/leaderboards/${gameId}/level/${level.id}/${cat.id}?top=1`,
        )

        if (lbRes.ok) {
          const lbData = await lbRes.json()
          const runs = lbData.data?.runs || []

          if (runs.length > 0 && runs[0].run) {
            const topRun = runs[0].run
            const primarySeconds = topRun.times?.primary_t

            if (typeof primarySeconds === "number") {
              const timeMs = Math.round(primarySeconds * 1000)
              worldRecords.push({
                levelName: level.name,
                levelId: level.id,
                categoryName: cat.name,
                categoryId: cat.id,
                timeMs: timeMs,
                runId: topRun.id,
                player: topRun.players?.[0]?.id || "Anonymous",
              })
            }
          }
        }

        // Small delay to prevent hammering the API too fast
        await new Promise((resolve) => setTimeout(resolve, 150))
      } catch (err) {
        // Skip if a specific combo doesn't exist
      }
    }
  }

  console.log(
    `--- Successfully Fetched ${worldRecords.length} World Records ---`,
  )
  console.table(worldRecords)
  return worldRecords
}

getExactWorldRecords()
