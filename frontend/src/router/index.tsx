import React from 'react';
import { Router as ReactRouter, Route, Switch, Redirect } from 'react-router';
import { Paths } from './paths';
import { history } from './history';

import { BlockHolesPage } from "../pages/block-holes";
import { SearchIndexesPage } from "../pages/search-indexes";
import { DmeshPage } from "../pages/dmesh";

export function Router(): React.ReactElement {
  return (
    <ReactRouter history={history}>
      <Switch>
        <Route exact path={Paths.blocks} component={BlockHolesPage} />
        <Route exact path={Paths.indexes} component={SearchIndexesPage} />
        <Route exact path={Paths.dmesh} component={DmeshPage} />
        <Redirect to={Paths.blocks} />
      </Switch>
    </ReactRouter>
  );
}
