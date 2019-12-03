import React from 'react';
import { Router as ReactRouter, Route, Switch, Redirect } from 'react-router';
import { Paths } from './paths';
import { history } from './history';

import { BlockHolesPage } from "../pages/block-holes";
import { SearchIndexesPage } from "../pages/search-indexes";
import { DmeshPage } from "../pages/dmesh";
import { KvdbBlocksPage } from "../pages/kvdb-blocks"
import { KvdbTrxsPage } from "../pages/kvdb-trxs"

export function Router(): React.ReactElement {
  return (
    <ReactRouter history={history}>
      <Switch>
        <Route exact path={Paths.blocks} component={BlockHolesPage} />
        <Route exact path={Paths.indexes} component={SearchIndexesPage} />
        <Route exact path={Paths.dmesh} component={DmeshPage} />
        <Route exact path={Paths.kvdbBlocks} component={KvdbBlocksPage} />
        <Route exact path={Paths.kvdbTrx} component={KvdbTrxsPage} />
        <Redirect to={Paths.blocks} />
      </Switch>
    </ReactRouter>
  );
}
