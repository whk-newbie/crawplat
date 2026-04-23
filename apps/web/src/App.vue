<template>
  <div>
    <header>
      <nav>
        <router-link to="/login">Login</router-link>
        <router-link to="/projects">Projects</router-link>
        <router-link to="/spiders">Spiders</router-link>
        <router-link to="/schedules">Schedules</router-link>
        <router-link to="/executions">Executions</router-link>
        <router-link to="/monitor">Monitor</router-link>
        <router-link to="/datasources">Datasources</router-link>
      </nav>
    </header>
    <router-view />
  </div>
</template>
