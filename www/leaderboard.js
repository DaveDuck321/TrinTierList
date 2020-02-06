// {people:[{id:number, name:string, imgs:string[]}], rankings: }

import { PostJSON, PopulateCategories } from "/common.js";

function sortByCategory(people, elos, category) {
    for (const person of people)
        person.elo = elos[category][person.id];

    people.sort((p1, p2) => p2.elo - p1.elo);
}

async function PopulateLeaderboard(people, elos, category) {
    // Delete existing entries
    document.querySelectorAll("#leaderboard tr").forEach(e => e.remove());

    sortByCategory(people, elos, category);

    const leaderboard = document.getElementById("leaderboard");
    const template = leaderboard.querySelector("template");
    const row = template.content.querySelector("tr");

    for (const [index, person] of people.entries()) {
        const personRow = document.importNode(row, true);

        personRow.querySelector(".rank").innerText = index + 1;
        personRow.querySelector(".nickname").innerText = person.nickname;
        personRow.querySelector(".elo").innerText = person.elo;

        leaderboard.appendChild(personRow);
    }
}

document.addEventListener("DOMContentLoaded", async () => {
    const { people, categories, elos } = await PostJSON("/api/leaderboard", {});

    PopulateLeaderboard(people, elos, categories[0].id);
    PopulateCategories(categories, 
        e => PopulateLeaderboard(people, elos, e.srcElement.value));
});
